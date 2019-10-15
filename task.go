package conversion

import (
	"context"
	"encoding/json"
	"github.com/gocacher/cacher"
	"sync"
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
}

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
	log.With("id", walk.Walk().ID()).Info("add walk")
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
				wg.Add(1)
				go func(walk IWalk) {
					e = walk.Run(t.Context)
					if e != nil {
						log.With("id", walk.ID()).Error("run:", e)
					}
					wg.Done()
				}(walk)
				t.running.Delete(walk.ID())
				//time.Sleep(1 * time.Second)
			}
			//time.Sleep(1 * time.Second)
			continue
		}
		wg.Wait()
		break
	}
	log.Info("end")
	return nil
}

// NewTask ...
func NewTask() *Task {
	return &Task{
		Context: context.Background(),
		queue:   sync.Pool{},
	}
}
