package conversion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gocacher/cacher"
	"go.uber.org/atomic"
)

// Running ...
type Running struct {
	running sync.Map
	waiting sync.Map
}

// Task ...
type Task struct {
	context  context.Context
	cancel   context.CancelFunc
	running  Running
	queue    sync.Pool
	autoStop *atomic.Bool
	Limit    int
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

// PoolMessage ...
type PoolMessage map[string][]byte

// Add ...
func (r *Running) Add(s string) {
	r.waiting.Store(s, nil)
	if err := r.cache(); err != nil {
		log.Error(err)
	}
}

// Delete ...
func (r *Running) Delete(s string) {
	r.waiting.Delete(s)
	if err := r.cache(); err != nil {
		log.Error(err)
	}
}

// Running ...
func (r *Running) Running(s string) (b bool) {
	_, b = r.running.LoadOrStore(s, nil)
	return
}

// Finish ...
func (r *Running) Finish(s string) {
	r.running.Delete(s)
}

// Has ...
func (r *Running) Has(s string) (b bool) {
	_, b = r.running.Load(s)
	return
}

// List ...
func (r *Running) List() []string {
	var runs []string
	r.running.Range(func(key, value interface{}) bool {
		if v, b := key.(string); b {
			runs = append(runs, v)
		}
		return true
	})
	return runs
}

// Restore ...
func (r *Running) Restore() ([]string, error) {
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
	}
	return runs, nil
}

func (r *Running) cache() error {
	bytes, e := json.Marshal(r.List())
	if e != nil {
		return Wrap(e)
	}
	return Wrap(cacher.Set("running", bytes))
}

// AddWorker ...
func (t *Task) AddWorker(work IWork, force bool) error {
	log.With("id", work.ID()).Info("add work")
	iwork, e := LoadWork(work.ID())
	if e == nil {
		if force || iwork.Status() == WorkStopped {
			if err := iwork.Reset(); err != nil {
				return Wrap(err)
			}
		}
	}
	if err := work.Store(); err != nil {
		return err
	}

	t.queue.Put(work.ID())
	return nil
}

// Stop ...
func (t *Task) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
}

// restore ...
func (t *Task) restore() error {
	ss, e := t.running.Restore()
	if e != nil {
		return Wrap(e)
	}
	for k := range ss {
		t.queue.Put(k)
	}
	return nil
}

// Running ...
func (t *Task) Running(work IWork) (b bool) {
	_, b = t.running.LoadOrStore(work.ID(), nil)
	return
}

// Finish ...
func (t *Task) Finish(id string) {
	t.running.Delete(id)
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
		return Wrap(err)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < t.Limit; i++ {
		wg.Add(1)
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

				if v := t.queue.Get(); v != nil {
					if s, b := v.(string); b {
						work, e := LoadWork(s)
						if e != nil {
							log.With("id", s, "error", e).Error("load work")
							continue
						}

						if t.Running(work) {
							log.With("id", work.ID()).Warn("work was running")
							return
						}

						//move to add
						//if work.Status() == WorkRunning {
						//	e := work.Reset()
						//	if e != nil {
						//		log.With("id", work.ID(), "error", e).Error("reset")
						//		continue
						//	}
						//}
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
						e = DeleteTaskMessage(work.ID())
						if e != nil {
							log.With("id", work.ID(), "error", e).Error("before run")
						}
						t.Finish(work.ID())
					}
					continue
				}
				if t.AutoStop() {
					break WorkEnd
				}
				//service waiting for new Work
				time.Sleep(30 * time.Second)
			}
		}(&wg)
	}
	wg.Wait()
	log.Info("end")
	return nil
}

// GetWorkStatus ...
func (t *Task) GetWorkStatus(id string) (WorkStatus, error) {
	work, e := LoadWork(id)
	if e != nil {
		return WorkAbnormal, fmt.Errorf("get status:%w", e)
	}
	_, ok := t.running.Load(work.ID())
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
	if err := iwork.Reset(); err != nil {
		return Wrap(err)
	}
	if err := t.AddWorker(iwork); err != nil {
		return Wrap(err)
	}
	return nil
}

// StopWork ...
func (t *Task) StopWork(id string) {
	if value, ok := t.running.Load(id); ok {
		if work, b := value.(IWork); b {
			if e := work.Stop(); e != nil {
				log.Error(e)
			} else {
				return
			}
		}
	}
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
	t.running.Range(func(key, value interface{}) bool {
		iwork, err := LoadWork(key.(string))
		if err != nil {
			e = Wrap(err)
			return false
		}
		works = append(works, iwork)
		return true
	})
	return
}

// NewTask ...
func NewTask() *Task {
	ctx, cancel := context.WithCancel(context.Background())
	return &Task{
		context:  ctx,
		cancel:   cancel,
		running:  sync.Map{},
		queue:    sync.Pool{},
		autoStop: atomic.NewBool(true),
		Limit:    DefaultLimit,
	}
}
