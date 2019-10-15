package conversion

import (
	"encoding/json"
	"sync"

	"github.com/go-cacher/cacher"
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
	Limit atomic.Int32
	queue sync.Pool
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
	for v := t.queue.Get(); v != nil {

	}

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
