package conversion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	cache "github.com/gocacher/badger-cache"
	"github.com/gocacher/cacher"
	"go.uber.org/atomic"
)

// Queue ...
type Queue struct {
	cacher  cacher.Cacher
	queuing *sync.Map
	tasking *sync.Pool
	running *sync.Map
}

// Task ...
type Task struct {
	context   context.Context
	cancel    context.CancelFunc
	queue     *Queue
	autoStop  *atomic.Bool
	Limit     int
	Interval  int
	ClearTemp bool
}

// AutoStop ...
func (t *Task) AutoStop() bool {
	return t.autoStop.Load()
}

// SetAutoStop ...
func (t *Task) SetAutoStop(autoStop bool) {
	t.autoStop.Store(autoStop)
}

// DefaultLimit ...
var DefaultLimit = 3

// NewQueue ...
func NewQueue(c cacher.Cacher) *Queue {
	return &Queue{
		cacher:  c,
		queuing: &sync.Map{},
		tasking: &sync.Pool{},
		running: &sync.Map{},
	}
}

// Add ...
func (q *Queue) Add(s string) {
	q.queuing.Store(s, nil)
	if err := q.cache(); err != nil {
		log.Error(err)
	}
	q.tasking.Put(s)
}

// Delete ...
func (q *Queue) Delete(s string) {
	q.queuing.Delete(s)
	if err := q.cache(); err != nil {
		log.Error(err)
	}
}

// Get ...
func (q *Queue) Get() (v string, b bool) {
	if t := q.tasking.Get(); t != nil {
		if v, b = t.(string); b {
			if !q.Has(v) {
				return "", false
			}
		}

	}
	return
}

// Has ...
func (q *Queue) Has(s string) bool {
	_, ok := q.queuing.Load(s)
	return ok
}

// Running ...
func (q *Queue) Running(work IWork) (b bool) {
	_, b = q.running.LoadOrStore(work.ID(), work)
	return
}

// Stop ...
func (q *Queue) Stop(id string) {
	if v, b := q.running.Load(id); b {
		if work, b := v.(IWork); b {
			if e := work.Stop(); e != nil {
				log.Error(e)
			} else {
				return
			}
		}
	}
	q.Delete(id)
}

// Finish ...
func (q *Queue) Finish(s string) {
	q.running.Delete(s)
	q.Delete(s)
}

// IsRunning ...
func (q *Queue) IsRunning(s string) (b bool) {
	_, b = q.running.Load(s)
	return
}

// List ...
func (q *Queue) List() []string {
	var runs []string
	q.queuing.Range(func(key, value interface{}) bool {
		if v, b := key.(string); b {
			runs = append(runs, v)
		}
		return true
	})
	return runs
}

// Restore ...
func (q *Queue) Restore() ([]string, error) {
	bytes, e := cacher.Get("running")
	if e != nil {
		return nil, e
	}
	var runs []string
	e = json.Unmarshal(bytes, &runs)
	if e != nil {
		return nil, e
	}
	for _, run := range runs {
		work, err := LoadWork(run)
		if err != nil {
			return nil, err
		}
		err = work.Reset()
		if err != nil {
			return nil, err
		}
		q.Add(run)
	}
	return runs, nil
}

func (q *Queue) cache() error {
	bytes, e := json.Marshal(q.List())
	if e != nil {
		return Wrap(e)
	}
	return Wrap(cacher.Set("running", bytes))
}

// AddWorker ...
func (t *Task) AddWorker(work IWork, force bool) error {
	log.With("id", work.ID()).Info("add work")
	if t.queue.IsRunning(work.ID()) {
		return nil
	}

	iwork, e := LoadWork(work.ID())
	if e == nil {
		if force {
			if err := iwork.Reset(); err != nil {
				return err
			}
		}
	} else {
		if err := work.Store(); err != nil {
			return err
		}
	}
	return t.StartWork(work.ID())
}

// Stop ...
func (t *Task) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
}

// restore ...
func (t *Task) restore() error {
	ss, e := t.queue.Restore()
	if e != nil {
		return Wrap(e)
	}
	for _, v := range ss {
		t.queue.Add(v)
	}
	return nil
}

// Start ...
func (t *Task) Start() error {
	if !CheckDatabase() {
		return errors.New("sql service was not ready")
	}
	if !CheckNode() {
		return errors.New("node service was not ready")
	}

	if err := t.restore(); err != nil {
		//ignore restore:first error key not found
		log.Warnw("if not your first run,this has some problems", "error", err)
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < t.Limit; i++ {
		wg.Add(1)
		log.Infow("task thread start", "idx", i)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
		WorkEnd:
			for {
				select {
				case <-t.context.Done():
					log.With("error", t.context.Err()).Error("done")
					return
				default:
				}
				if v, b := t.queue.Get(); b {
					work, e := LoadWork(v)
					if e != nil {
						log.With("id", v, "error", e).Error("load work")
						continue
					}

					if t.queue.Running(work) {
						log.With("id", work.ID()).Warn("work was running")
						continue
					}

					switch work.Status() {
					case WorkWaiting:
						log.With("id", work.ID()).Info("work run")

						e = work.Run(t.context)
						if e != nil {
							log.With("id", work.ID(), "error", e).Error("run")
						}
					case WorkStopped:
						log.With("id", work.ID()).Info("work was stopped")
						continue
					case WorkRunning:
						log.With("id", work.ID()).Warn("work was running")
						continue
					case WorkFinish:
						log.With("id", work.ID()).Warn("work was finished")
						continue
					default:
						log.With("id", work.ID()).Error("work status wrong")
						e := work.Reset()
						if e != nil {
							log.With("id", work.ID(), "error", e).Error("fix status error")
							return
						}
					}
					log.With("id", work.ID()).Info("end run")
					t.queue.Finish(work.ID())
					continue
				}
				if t.AutoStop() {
					break WorkEnd
				}
				//service queuing for new Work
				time.Sleep(5 * time.Second)
			}
		}(wg)
	}

	log.Info("waiting for end")
	wg.Wait()
	return nil
}

// GetWorkStatus ...
func (t *Task) GetWorkStatus(id string) (WorkStatus, error) {
	work, e := LoadWork(id)
	if e != nil {
		return WorkAbnormal, fmt.Errorf("get status:%w", e)
	}
	ok := t.queue.Has(work.ID())
	if !ok && work.Status() == WorkRunning {
		return WorkWaiting, nil
	}
	return work.Status(), nil
}

// GetWork ...
func (t *Task) GetWork(id string) (IWork, error) {
	return LoadWork(id)
}

// StartWork ...
func (t *Task) StartWork(id string) error {
	iwork, e := LoadWork(id)
	if e != nil {
		return Wrap(e)
	}
	if iwork.Status() == WorkStopped {
		if err := iwork.Reset(); err != nil {
			return Wrap(err)
		}
	}
	t.queue.Add(iwork.ID())
	return nil
}

// StopWork ...
func (t *Task) StopWork(id string) {
	//stop running
	t.queue.Stop(id)

	//change stop status
	iwork, err := LoadWork(id)
	if err == nil {
		if e := iwork.Stop(); e != nil {
			log.Error(e)
		} else {
			return
		}
	} else {
		log.Error(err)
	}
}

// AllRun ...
func (t *Task) AllRun() (works []IWork, e error) {
	for _, v := range t.queue.List() {
		iwork, err := LoadWork(v)
		if err != nil {
			return nil, err
		}
		works = append(works, iwork)
	}
	return works, nil
}

// NewTask ...
func NewTask() *Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Task{
		context:  ctx,
		cancel:   cancel,
		queue:    NewQueue(cache.NewBadgerCache(CachePath)),
		autoStop: atomic.NewBool(true),
		Limit:    DefaultLimit,
	}
}
