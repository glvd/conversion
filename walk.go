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
	WalkImpl
	VideoPath  string
	PosterPath string
	ThumbPath  string
	SamplePath []string
	Scale      Scale
	Output     string
	Skip       []string
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
type VideoProcessFunc func(src []byte) (IVideo, error)

// WalkOptions ...
type WalkOptions func(walk *Walk)

// ErrWalkFinish ...
var ErrWalkFinish = errors.New("walk was finished")

// ErrWrongCastType ...
var ErrWrongCastType = errors.New("something wrong when cast to type")

// WalkRunProcessFunction ...
var WalkRunProcessFunction = map[string]VideoProcessFunc{
	"source": decodeSource,
	"info":   decodeInfo,
}

// SkipOption ...
func SkipOption(skip ...string) WalkOptions {
	return func(walk *Walk) {
		walk.Skip = skip
	}
}

// ScaleOption ...
func ScaleOption(scale Scale) WalkOptions {
	return func(walk *Walk) {
		walk.Scale = scale
	}
}

// OutputPathOption ...
func OutputPathOption(path string) WalkOptions {
	return func(walk *Walk) {
		walk.Output = path
	}
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

func dummy(src []byte) (IVideo, error) {
	log.With("src", string(src)).Panic("dummy")
	return nil, nil
}

// Reset ...
func (w *Walk) Reset() error {
	w.WalkImpl.Status = WalkWaiting
	return w.Update()
}

// Status ...
func (w *Walk) Status() WalkStatus {
	return w.WalkImpl.Status
}

// Walk ...
func (w Walk) Walk() Walk {
	return w
}

func (w Walk) slice() *Slice {
	return NewSlice(w.VideoPath, SliceScale(w.Scale), SliceOutput(w.Output), SliceSkip(w.Skip...))
}

func (w Walk) video() (IVideo, error) {
	fn := dummy
	fn, b := WalkRunProcessFunction[w.WalkType]
	if !b {
		fn = dummy
	}
	//time.Sleep(5 * time.Second)
	return fn(w.Value)
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
	v, e := w.video()
	if e != nil {
		return e
	}
	video := v.Video()
	i, e := InsertOrUpdate(video)
	if e != nil {
		return e
	}
	if i == 0 {
		log.With("id", video.ID).Warn("not updated")
	}
	e = w.slice().Do(ctx)
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
