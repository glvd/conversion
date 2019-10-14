package conversion

// RunType ...
type RunType string

// RunTypePath ...
const (
	RunTypePath RunType = "path"
	RunTypeJSON RunType = "json"
)

// Task ...
type Task struct {
}

// AddWalker ...
func (t Task) AddWalker(walker IWalker) {

}
