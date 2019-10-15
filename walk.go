package conversion

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gocacher/cacher"
	"time"
)

// WalkRunning ...
const (
	WalkWaiting WalkStatus = iota + 1
	WalkRunning
	WalkFinish
)

// WalkStatus ...
type WalkStatus int

// WalkImpl ...
type WalkImpl struct {
	ID       string
	WalkType string
	Status   WalkStatus
	Value    []byte
}

// Walk ...
type Walk struct {
	WalkImpl
}

// IWalk ...
type IWalk interface {
	ID() string
	Walk() Walk
	Update() error
	Store() error
	Reset() error
	Status() WalkStatus
	Run(ctx context.Context) (e error)
}

// ErrWalkFinish ...
var ErrWalkFinish = errors.New("walk was finished")

// ErrWrongCastType ...
var ErrWrongCastType = errors.New("something wrong when cast to type")

// VideoProcessFunc ...
type VideoProcessFunc func(src []byte) error

// WalkRunProcessFunction ...
var WalkRunProcessFunction = map[string]VideoProcessFunc{
	"source": SourceProcess,
	"info":   nil,
}

func dummy(v []byte) error {
	log.Panic(v)
	return nil
}

// Reset ...
func (w *Walk) Reset() error {
	w.WalkImpl.Status = WalkWaiting
	return w.Store()
}

func (w *Walk) Status() WalkStatus {
	return w.WalkImpl.Status
}

// Walk ...
func (w Walk) Walk() Walk {
	return w
}

// LoadWalk ...
func LoadWalk(id string) (IWalk, error) {
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
	return w.WalkImpl.ID
}

// Store ...
func (w *Walk) Store() error {
	bytes, e := json.Marshal(w)
	if e != nil {
		return e
	}
	b, err := cacher.Has(w.ID())
	if err != nil {
		return err
	}
	if b {
		log.With("id", w.ID()).Warn("store")
		return nil
	}
	return cacher.Set(w.ID(), bytes)
}

func (w *Walk) Update() error {
	bytes, e := json.Marshal(w)
	if e != nil {
		return e
	}
	return cacher.Set(w.ID(), bytes)
}

// Run ...
func (w *Walk) Run(ctx context.Context) (e error) {
	switch w.WalkImpl.Status {
	case WalkFinish:
		log.With("id", w.ID()).Warn("walk was finished")
		return nil
	case WalkRunning:
		log.With("id", w.ID()).Warn("walk was running")
		return nil
	case WalkWaiting:
		w.WalkImpl.Status = WalkRunning
		if err := w.Update(); err != nil {
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
	time.Sleep(5 * time.Second)
	e = fn(w.Value)
	if e != nil {
		return e
	}

	w.WalkImpl.Status = WalkFinish
	if err := w.Update(); err != nil {
		return err
	}
	return nil
}
