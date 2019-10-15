package conversion

import "sync"

// RunType ...
type RunType string

// RunTypePath ...
const (
	RunTypePath RunType = "path"
	RunTypeJSON RunType = "json"
)

// Task ...
type Task struct {
	sync.Pool
}

// AddWalker ...
func (t Task) AddWalker(walk IWalk) {

}
