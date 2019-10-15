package conversion

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/go-cacher/cacher"
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
	bytes, e := cacher.Get("task")
	if e != nil {
		panic(e)
	}
	var s []string
	e = json.Unmarshal(bytes, &s)
	if e != nil {
		return nil, e
	}
	return s, nil
}

// AddWalker ...
func (t *Task) AddWalker(walk IWalk) error {
	if err := walk.Store(); err != nil {
		return err
	}
	t.queue.Put(walk.Walk().ID)
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
	for {
		if v := t.queue.Get(); v != nil {
			if s, b := v.(string); b {
				walk, e := LoadWalk(s)
				if e != nil {
					log.Error(e)
					continue
				}
				e = walk.Run(t.Context)
				if e != nil {
					log.Error(e)
				}
			}
		}
	}

}

// NewTask ...
func NewTask() *Task {
	return &Task{
		Context: context.Background(),
		queue:   sync.Pool{},
	}
}
