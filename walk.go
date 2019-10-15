package conversion

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-cacher/cacher"
)

// WalkRunning ...
const (
	WalkWaiting WalkStatus = iota + 1
	WalkRunning
	WalkFinish
)

// WalkStatus ...
type WalkStatus int

// Walk ...
type walk struct {
	ID       string
	WalkType string
	Status   WalkStatus
	Value    interface{}
}

// Walk ...
type Walk struct {
	walk
}

// IWalk ...
type IWalk interface {
	ID() string
	Walk() Walk
	Store() error
	Reset() error
	Run(ctx context.Context) (e error)
}

// ErrWalkFinish ...
var ErrWalkFinish = errors.New("walk was finished")

// ErrWrongCastType ...
var ErrWrongCastType = errors.New("something wrong when cast to type")

// VideoProcessFunc ...
type VideoProcessFunc func(v interface{}) error

// WalkRunProcessFunction ...
var WalkRunProcessFunction = map[string]VideoProcessFunc{
	"source": SourceProcess,
	"info":   nil,
}

func dummy(v interface{}) error {
	log.Panic(v)
	return nil
}

// Reset ...
func (w *Walk) Reset() error {
	w.Status = WalkWaiting
	return w.Store()
}

// Walk ...
func (w Walk) Walk() Walk {
	return w
}

// LoadWalk ...
func LoadWalk(id string) (*Walk, error) {
	bytes, e := cacher.Get(id)
	if e != nil {
		return nil, e
	}
	var w Walk
	e = json.Unmarshal(bytes, &w)
	return &w, e
}

// ID ...
func (w Walk) ID() string {
	return w.walk.ID
}

// Store ...
func (w *Walk) Store() error {
	bytes, e := json.Marshal(w)
	if e != nil {
		return e
	}
	return cacher.Set(w.ID(), bytes)
}

// Run ...
func (w *Walk) Run(ctx context.Context) (e error) {
	switch w.walk.Status {
	case WalkFinish:
		log.With("id", w.ID()).Warn("walk was finished")
		return nil
	case WalkRunning:
		log.With("id", w.ID()).Warn("walk was running")
		return nil
	case WalkWaiting:
		w.walk.Status = WalkRunning
		if err := w.Store(); err != nil {
			return err
		}
	default:
		log.With("id", w.ID()).Panic("walk status wrong")
	}
	fn := dummy
	fn, b := WalkRunProcessFunction[w.WalkType]
	if !b {
		fn = dummy
	}
	return fn(w.Value)
}
