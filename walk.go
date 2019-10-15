package conversion

import "context"

// WalkStatus ...
type WalkStatus int

// WalkRunning ...
const (
	WalkWaiting WalkStatus = iota + 1
	WalkRunning
	WalkFinish
)

// Walk ...
type Walk struct {
	ID     string
	Status WalkStatus
	Value  interface{}
}

// IWalk ...
type IWalk interface {
	LoadWalk() Walk
	Run(ctx context.Context) (e error)
}
