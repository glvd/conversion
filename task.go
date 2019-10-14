package conversion

// RunType ...
type RunType string

// RunTypePath ...
const (
	RunTypePath         = RunType("path")
	RunTypeJSON RunType = "json"
)

// Task ...
type Task struct {
	Name RunType
}

// RunTaskWithPath ...
func RunTaskWithPath(path string) {
}

// RunTaskWithJSON ...
func RunTaskWithJSON(file string) {

}
