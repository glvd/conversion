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
func (t Task) AutoStop() bool {
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

// AddWalker ...
func (t *Task) AddWalker(walk IWalk) error {
	log.With("id", walk.ID()).Info("add walk")
	if err := walk.Store(); err != nil {
		return err
	}
	e := AddTaskMessage(walk.ID())
	if e != nil {
		return Wrap(e)
	}
	t.queue.Put(walk.ID())
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
		WalkEnd:
			for {
				select {
				case <-t.context.Done():
					log.With("error", t.context.Err()).Error("done")
					return
				default:
				}
				if v := t.queue.Get(); v != nil {
					if s, b := v.(string); b {
						walk, e := LoadWalk(s)
						if e != nil {
							log.Error(e)
							continue
						}
						_, b := t.running.LoadOrStore(walk.ID(), nil)
						if !b && walk.Status() == WalkRunning {
							log.With("id", walk.ID()).Warn("reset status")
							e := walk.Reset()
							if e != nil {
								log.With("id", walk.ID()).Error("reset:", e)
							}
						}
						log.With("id", walk.ID()).Info("queue")
						switch walk.Status() {
						case WalkFinish:
							log.With("id", walk.ID()).Warn("walk was finished")
							continue
						case WalkRunning:
							log.With("id", walk.ID()).Warn("walk was running")
							continue
						case WalkWaiting:
							e := DeleteTaskMessage(walk.ID())
							if e != nil {
								log.With("id", walk.ID()).Error("before run:", e)
							}
							e = walk.Run(t.context)
							if e != nil {
								log.With("id", walk.ID()).Error("run:", e)
							}
						default:
							log.With("id", walk.ID()).Panic("walk status wrong")
							continue
						}
						log.With("id", walk.ID()).Info("run end")
						t.running.Delete(walk.ID())
					}
					continue
				}
				if t.AutoStop() {
					break WalkEnd
				}
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
