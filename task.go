package conversion

import (
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
	walk  map[string]IWalk
	queue sync.Pool
}

// AddWalker ...
func (t *Task) AddWalker(walk IWalk) {
}

// Start ...
func (t *Task) Start() {

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
