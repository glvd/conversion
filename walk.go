package conversion

import "context"

// IWalk ...
type IWalk interface {
	Run(ctx context.Context) (e error)
}
