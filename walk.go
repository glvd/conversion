package conversion

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocacher/cacher"
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
	VideoPath  string
	PosterPath string
	ThumbPath  string
	SamplePath []string
	WalkImpl
}

// IWalk ...
type IWalk interface {
	ID() string
	Update() error
	Store() error
	Reset() error
	Status() WalkStatus
	Run(ctx context.Context) (e error)
}

// VideoProcessFunc ...
type VideoProcessFunc func(walk *Walk) error

// WalkOptions ...
type WalkOptions func(walk *Walk)

// ErrWalkFinish ...
var ErrWalkFinish = errors.New("walk was finished")

// ErrWrongCastType ...
var ErrWrongCastType = errors.New("something wrong when cast to type")

// WalkRunProcessFunction ...
var WalkRunProcessFunction = map[string]VideoProcessFunc{
	"source": SourceProcess,
	"info":   InfoProcess,
}

// VideoPathOption ...
func VideoPathOption(path string) WalkOptions {
	return func(walk *Walk) {
		walk.VideoPath = path
	}
}

// ThumbPathOption ...
func ThumbPathOption(path string) WalkOptions {
	return func(walk *Walk) {
		walk.ThumbPath = path
	}
}

// PosterPathOption ...
func PosterPathOption(path string) WalkOptions {
	return func(walk *Walk) {
		walk.PosterPath = path
	}
}

// SamplePathOption ...
func SamplePathOption(path []string) WalkOptions {
	return func(walk *Walk) {
		walk.SamplePath = path
	}
}

func dummy(walk *Walk) error {
	log.With("id", walk.ID()).Panic(walk)
	return nil
}

// Reset ...
func (w *Walk) Reset() error {
	w.WalkImpl.Status = WalkWaiting
	return w.Store()
}

// Status ...
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

// Update ...
func (w *Walk) Update() error {
	bytes, e := json.Marshal(w)
	if e != nil {
		return e
	}
	b, err := cacher.Has(w.ID())
	if err != nil {
		return err
	}
	if !b {
		log.With("id", w.ID()).Warn("update")
		return nil
	}
	return cacher.Set(w.ID(), bytes)
}

// Run ...
func (w *Walk) Run(ctx context.Context) (e error) {
	w.WalkImpl.Status = WalkRunning
	if err := w.Update(); err != nil {
		return err
	}
	fn := dummy
	fn, b := WalkRunProcessFunction[w.WalkType]
	if !b {
		fn = dummy
	}
	//time.Sleep(5 * time.Second)
	e = fn(w)
	if e != nil {
		return e
	}

	w.WalkImpl.Status = WalkFinish
	return w.Update()
}

// GetFiles ...
func GetFiles(name string, regex string) (files []string) {
	info, e := os.Stat(name)
	if e != nil {
		return
	}
	if !info.IsDir() {
		return append(files, name)
	}
	file, e := os.Open(name)
	if e != nil {
		return
	}
	defer file.Close()
	names, e := file.Readdirnames(-1)
	if e != nil {
		return
	}
	var fullPath string
	for _, filename := range names {
		fullPath = filepath.Join(name, filename)
		base := filepath.Base(fullPath)
		if regex != "" && strings.Index(base, regex) == -1 {
			continue
		}
		fileInfo, e := os.Stat(fullPath)
		if e != nil || fileInfo.IsDir() {
			log.With("dir", fileInfo != nil).Error(e)
			continue
		}
		files = append(files, fullPath)
	}
	return files
}

// SkipVerify ...
func SkipVerify(tp string, v ...interface{}) bool {
	for i := range v {
		if v1, b := (v[i]).(string); b {
			if v1 == tp {
				return true
			}
		}
	}
	return false
}
