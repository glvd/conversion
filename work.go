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

// WorkRunning ...
const (
	WorkWaiting WorkStatus = iota + 1
	WorkRunning
	WorkFinish
)

// RelateList ...
const relateList = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// WorkStatus ...
type WorkStatus int

// WorkImpl ...
type WorkImpl struct {
	ID         string
	Status     WorkStatus
	VideoPaths []string
	PosterPath string
	ThumbPath  string
	SamplePath []string
	Scale      Scale
	Output     string
	Skip       []string
}

// Work ...
type Work struct {
	*WorkImpl
	WorkType string
	Value    []byte
}

// IWork ...
type IWork interface {
	ID() string
	Update() error
	Store() error
	Reset() error
	Status() WorkStatus
	Run(ctx context.Context) (e error)
}

// VideoProcessFunc ...
type VideoProcessFunc func(src []byte) (IVideo, error)

// WorkOptions ...
type WorkOptions func(impl *WorkImpl)

// ErrWorkFinish ...
var ErrWorkFinish = errors.New("Work was finished")

// ErrWrongCastType ...
var ErrWrongCastType = errors.New("something wrong when cast to type")

// WorkRunProcessFunction ...
var WorkRunProcessFunction = map[string]VideoProcessFunc{
	"source": decodeSource,
	"info":   decodeInfo,
}

func IDOption(id string) WorkOptions {
	return func(impl *WorkImpl) {
		impl.ID = id
	}
}

// SkipOption ...
func SkipOption(skip ...string) WorkOptions {
	return func(impl *WorkImpl) {
		impl.Skip = skip
	}
}

// ScaleOption ...
func ScaleOption(scale Scale) WorkOptions {
	return func(impl *WorkImpl) {
		impl.Scale = scale
	}
}

// OutputPathOption ...
func OutputPathOption(path string) WorkOptions {
	return func(impl *WorkImpl) {
		impl.Output = path
	}
}

// VideoPathOption ...
func VideoPathOption(path []string) WorkOptions {
	return func(impl *WorkImpl) {
		impl.VideoPaths = path
	}
}

// ThumbPathOption ...
func ThumbPathOption(path string) WorkOptions {
	return func(impl *WorkImpl) {
		impl.ThumbPath = path
	}
}

// PosterPathOption ...
func PosterPathOption(path string) WorkOptions {
	return func(impl *WorkImpl) {
		impl.PosterPath = path
	}
}

// SamplePathOption ...
func SamplePathOption(path []string) WorkOptions {
	return func(impl *WorkImpl) {
		impl.SamplePath = path
	}
}

func dummy(src []byte) (IVideo, error) {
	log.With("src", string(src)).Panic("dummy")
	return nil, nil
}

func defaultWork(options ...WorkOptions) *WorkImpl {
	impl := &WorkImpl{
		ID:         "",
		Status:     0,
		VideoPaths: nil,
		PosterPath: "",
		ThumbPath:  "",
		SamplePath: nil,
		Scale:      0,
		Output:     os.TempDir(),
		Skip:       nil,
	}
	for _, opt := range options {
		opt(impl)
	}
	return impl
}

// Reset ...
func (w *Work) Reset() error {
	w.WorkImpl.Status = WorkWaiting
	return w.Update()
}

// Status ...
func (w *Work) Status() WorkStatus {
	return w.WorkImpl.Status
}

func (w Work) slice(ctx context.Context, input string) (*Fragment, error) {
	format, e := split.FFProbeStreamFormat(input)
	if e != nil {
		return nil, Wrap(e)
	}
	if !IsMedia(format) {
		return nil, errors.New("file is not a video/audio")
	}
	res := parseScale(int64(format.ResolutionInt()))
	if res < w.Scale {
		w.Scale = res
	}
	sharpness := strconv.FormatInt(formatScale(w.Scale), 10) + "P"

	//Output := filepath.Join(w.Output, UUID().String())
	sa, e := split.FFMpegSplitToM3U8(ctx, input, split.StreamFormatOption(format), split.ScaleOption(formatScale(w.Scale)), split.OutputOption(w.Output()), split.AutoOption(true))
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

func (w Work) video() (IVideo, error) {
	fn, b := WorkRunProcessFunction[w.WorkType]
	if !b {
		fn = dummy
	}
	return fn(w.Value)
}

// LoadWork ...
func LoadWork(id string) (IWork, error) {
	bytes, e := cacher.Get(id)
	if e != nil {
		return nil, e
	}
	var w Work
	e = json.Unmarshal(bytes, &w)
	return &w, e
}

// ID ...
func (w Work) ID() string {
	return w.WorkImpl.ID
}

func (w Work) Output() string {
	return w.WorkImpl.Output
}

// Store ...
func (w *Work) Store() error {
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
func (w *Work) Update() error {
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
func (w *Work) Run(ctx context.Context) (e error) {
	w.WorkImpl.Status = WorkRunning
	if err := w.Update(); err != nil {
		return Wrap(err)
	}
	v, e := w.video()
	if e != nil {
		return Wrap(e)
	}
	for _, path := range w.VideoPaths {
		if path == "" {
			continue
		}
		video := v.Video()
		video.TotalEpisode = strconv.Itoa(len(w.VideoPaths))
		video.Episode = strconv.Itoa(GetFileIndex(path))
		if !ExistVerifyString("source", w.Skip...) {
			s, e := AddFile(ctx, path)
			if e != nil {
				return Wrap(e)
			}
			video.SourceHash = s
		}

		if !ExistVerifyString("slice", w.Skip...) {
			f, e := w.slice(ctx, path)
			if e != nil {
				return Wrap(e)
			}
			s, e := AddDir(ctx, f.Output())
			if e != nil {
				return Wrap(e)
			}
			video.M3U8Hash = s
			//AddHash(s)
		}
		if !ExistVerifyString("poster", w.Skip...) && w.PosterPath != "" {
			s, e := AddFile(ctx, w.PosterPath)
			if e != nil {
				return Wrap(e)
			}
			video.PosterHash = s
		}

		if !ExistVerifyString("thumb", w.Skip...) && w.PosterPath != "" {
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

	w.WorkImpl.Status = WorkFinish
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
