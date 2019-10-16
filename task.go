package conversion

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gocacher/cacher"
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
	Context context.Context
	running sync.Map
	queue   sync.Pool
	Limit   int
}

// DefaultLimit ...
var DefaultLimit = 5

// StoreTask ...
func StoreTask(s []string) error {
	bytes, e := json.Marshal(s)
	if e != nil {
		return e
	}
	return cacher.Set("task", bytes)
}

// LoadTask ...
func LoadTask() ([]string, error) {
	b, e := cacher.Has("task")
	if e != nil {
		return nil, e
	}
	var s []string
	if b {
		bytes, e := cacher.Get("task")
		if e != nil {
			return nil, e
		}
		e = json.Unmarshal(bytes, &s)
		if e != nil {
			return nil, e
		}
	}
	return s, nil
}

// AddWalker ...
func (t *Task) AddWalker(walk IWalk) error {
	log.With("id", walk.ID()).Info("add walk")
	if err := walk.Store(); err != nil {
		return err
	}
	t.queue.Put(walk.ID())
	return nil
}

// Start ...
func (t *Task) Start() error {
	ss, e := LoadTask()
	if e != nil {
		return e
	}
	for _, s := range ss {
		t.queue.Put(s)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < t.Limit; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for {
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
							e = walk.Run(t.Context)
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
				break
			}
		}(&wg)
	}
	wg.Wait()
	log.Info("end")
	return nil
}

// NewTask ...
func NewTask() *Task {
	return &Task{
		Context: context.Background(),
		Limit:   DefaultLimit,
		queue:   sync.Pool{},
	}
}
