package conversion

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gocacher/cacher"
	"go.uber.org/atomic"
)

// RunType ...
type RunType string

// RunTypePath ...
const (
	RunTypePath RunType = "path"
	RunTypeJSON RunType = "json"
)

// Task ...
type Task struct {
	context  context.Context
	cancel   context.CancelFunc
	running  sync.Map
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

// AddTaskMessage ...
func AddTaskMessage(s string) error {
	messages, e := LoadTaskMessage()
	if e != nil {
		return e
	}
	messages[s] = nil
	bytes, e := json.Marshal(messages)
	if e != nil {
		return Wrap(e)
	}
	return Wrap(cacher.Set("task", bytes))
}

// DeleteTaskMessage ...
func DeleteTaskMessage(s string) error {
	messages, e := LoadTaskMessage()
	if e != nil {
		return Wrap(e)
	}
	delete(messages, s)
	bytes, e := json.Marshal(messages)
	if e != nil {
		return Wrap(e)
	}
	return Wrap(cacher.Set("task", bytes))
}

// LoadTaskMessage ...
func LoadTaskMessage() (PoolMessage, error) {
	b, e := cacher.Has("task")
	if e != nil {
		return nil, Wrap(e)
	}
	msg := make(PoolMessage)
	if b {
		bytes, e := cacher.Get("task")
		if e != nil {
			return nil, Wrap(e)
		}
		e = json.Unmarshal(bytes, &msg)
		if e != nil {
			return nil, Wrap(e)
		}
	}
	return msg, nil
}

// AddWorker ...
func (t *Task) AddWorker(Work IWork) error {
	log.With("id", Work.ID()).Info("add Work")
	if err := Work.Store(); err != nil {
		return err
	}
	e := AddTaskMessage(Work.ID())
	if e != nil {
		return Wrap(e)
	}
	t.queue.Put(Work.ID())
	return nil
}

// Stop ...
func (t *Task) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
}

// Restore ...
func (t *Task) Restore() error {
	ss, e := LoadTaskMessage()
	if e != nil {
		return Wrap(e)
	}
	for k := range ss {
		t.queue.Put(k)
	}
	return nil
}

// IsRunning ...
func (t *Task) IsRunning(id string) (b bool) {
	_, b = t.running.LoadOrStore(id, nil)
	return
}

// Start ...
func (t *Task) Start() error {
	if !CheckDatabase() || !CheckNode() {
		return errors.New("service was not ready")
	}
	if err := t.Restore(); err != nil {
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
						Work, e := LoadWork(s)
						if e != nil {
							log.With("id", s, "error", e).Error("load Work")
							continue
						}
						if !t.IsRunning(Work.ID()) && Work.Status() == WorkRunning {
							e := Work.Reset()
							if e != nil {
								log.With("id", Work.ID(), "error", e).Error("reset")
								continue
							}
						}
						switch Work.Status() {
						case WorkFinish:
							log.With("id", Work.ID()).Warn("Work was finished")
							continue
						case WorkRunning:
							log.With("id", Work.ID()).Warn("Work was running")
							continue
						case WorkWaiting:
							log.With("id", Work.ID()).Info("Work run")
							e := DeleteTaskMessage(Work.ID())
							if e != nil {
								log.With("id", Work.ID(), "error", e).Error("before run")
							}
							e = Work.Run(t.context)
							if e != nil {
								log.With("id", Work.ID(), "error", e).Error("run")
							}
						default:
							log.With("id", Work.ID()).Panic("Work status wrong")
							continue
						}
						log.With("id", Work.ID()).Info("end run")
						t.running.Delete(Work.ID())
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
