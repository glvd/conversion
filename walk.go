package conversion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/glvd/split"

	"github.com/gocacher/cacher"
)

// WalkRunning ...
const (
	WalkWaiting WalkStatus = iota + 1
	WalkRunning
	WalkFinish
)

// RelateList ...
const relateList = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
	VideoPath  []string
	PosterPath string
	ThumbPath  string
	SamplePath []string
	Scale      Scale
	output     string
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
		walk.output = path
	}
}

// VideoPathOption ...
func VideoPathOption(path []string) WalkOptions {
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

// Output ...
func (w Walk) Output() string {
	if w.output == "" {
		return "tmp"
	}
	return w.output
}

// Walk ...
func (w Walk) Walk() Walk {
	return w
}
func (w Walk) slice(ctx context.Context, input string) (*Fragment, error) {
	format, e := split.FFProbeStreamFormat(input)
	if e != nil {
		return nil, Wrap(e)
	}
	if !IsMedia(format) {
		return nil, errors.New("file is not a video/audio")
	}
	res := toScale(int64(format.ResolutionInt()))
	if res < w.Scale {
		w.Scale = res
	}
	sharpness := strconv.FormatInt(scale(w.Scale), 10) + "P"

	//output := filepath.Join(w.Output, UUID().String())
	sa, e := split.FFMpegSplitToM3U8(ctx, input, split.StreamFormatOption(format), split.ScaleOption(scale(w.Scale)), split.OutputOption(w.Output()), split.AutoOption(true))
	if e != nil {
		return nil, Wrap(e)
	}
	return &Fragment{
		scale:     w.Scale,
		output:    sa.Output,
		skip:      w.Skip,
		input:     input,
		sharpness: sharpness,
	}, nil
}

func (w Walk) video() (IVideo, error) {
	fn, b := WalkRunProcessFunction[w.WalkType]
	if !b {
		fn = dummy
	}
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
		return Wrap(err)
	}
	v, e := w.video()
	if e != nil {
		return Wrap(e)
	}
	for _, path := range w.VideoPath {
		if path == "" {
			continue
		}
		video := v.Video()
		video.TotalEpisode = strconv.Itoa(len(w.VideoPath))
		video.Episode = strconv.Itoa(GetFileIndex(path))
		if !SkipVerifyString("source", w.Skip...) {
			s, e := AddFile(ctx, path)
			if e != nil {
				return Wrap(e)
			}
			video.SourceHash = s
		}

		if !SkipVerifyString("slice", w.Skip...) {
			f, e := w.slice(ctx, path)
			if e != nil {
				return Wrap(e)
			}
			s, e := AddDir(ctx, f.Output())
			if e != nil {
				return Wrap(e)
			}
			video.M3U8Hash = s
		}
		if !SkipVerifyString("poster", w.Skip...) && w.PosterPath != "" {
			s, e := AddFile(ctx, w.PosterPath)
			if e != nil {
				return Wrap(e)
			}
			video.PosterHash = s
		}

		if !SkipVerifyString("thumb", w.Skip...) && w.PosterPath != "" {
			s, e := AddFile(ctx, w.ThumbPath)
			if e != nil {
				return Wrap(e)
			}
			video.ThumbHash = s
		}

		i, e := InsertOrUpdate(video)
		if e != nil {
			return Wrap(e)
		}
		if i == 0 {
			log.With("id", video.ID()).Warn("not updated")
		}
	}

	w.WalkImpl.Status = WalkFinish
	return Wrap(w.Update())
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

// GetFileIndex ...
func GetFileIndex(filename string) int {
	return GetNameIndex(filepath.Base(filename))
}

// GetNameIndex ...
func GetNameIndex(name string) int {
	last := LastSplit(FileAbsName(name), "@")
	idx := ByteIndex(last[0])
	if ByteIndex(last[0]) == -1 {
		return 1
	}
	return idx + 1
}

// FileAbsName ...
func FileAbsName(filename string) string {
	_, filename = filepath.Split(filename)
	for i := len(filename) - 1; i >= 0 && !os.IsPathSeparator(filename[i]); i-- {
		if filename[i] == '.' {
			return filename[:i]
		}
	}
	return ""
}

// FileName ...
func FileName(filename string) string {
	s := []rune(FileAbsName(filename))
	last := len(s) - 1
	if last > 0 && unicode.IsLetter(s[last]) {
		if s[last-1] == '@' {
			return string(s[:last-1])
		}
	}
	return string(s)
}

// IndexByte ...
func IndexByte(index int) byte {
	if index > len(relateList) {
		return relateList[0]
	}
	return relateList[index]
}

// ByteIndex ...
func ByteIndex(idx byte) int {
	return strings.IndexByte(relateList, idx)
}

// LastSplit ...
func LastSplit(s, sep string) string {
	ss := strings.Split(s, sep)
	for i := len(ss) - 1; i >= 0; i-- {
		if ss[i] == "" {
			continue
		}
		return ss[i]
	}
	return ""
}

// Wrap ...
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w", err)
}
