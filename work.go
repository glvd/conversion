package conversion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/glvd/go-fftool"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/gocacher/cacher"
)

// WorkWaiting ...
const (
	WorkAbnormal WorkStatus = iota
	WorkWaiting
	WorkRunning
	WorkStopped
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
	Crypto     *Crypto
	Scale      Scale
	Output     string
	Skip       []string
	ClearTemp  bool
}

// Work ...
type Work struct {
	ctx    context.Context
	cancel context.CancelFunc
	*WorkImpl
	WorkType string
	Value    []byte
}

// IWork ...
type IWork interface {
	ID() string
	Work() *Work
	Video() *Video
	Info() string
	Update() error
	Store() error
	Reset() error
	Status() WorkStatus
	Run(ctx context.Context) (e error)
	Stop() error
}

// VideoProcessFunc ...
type VideoProcessFunc func(src []byte) (IVideo, error)

// WorkOptions ...
type WorkOptions func(impl *WorkImpl)

// ErrWorkFinish ...
var ErrWorkFinish = errors.New("work was finished")

// ErrWorkID ...
var ErrWorkID = errors.New("video id must input")

// ErrWrongCastType ...
var ErrWrongCastType = errors.New("something wrong when cast to type")

// WorkRunProcessFunction ...
var WorkRunProcessFunction = map[string]VideoProcessFunc{
	"source": decodeSource,
	"info":   decodeInfo,
}

// IDOption ...
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

// CryptoOption ...
func CryptoOption(crypto *Crypto) WorkOptions {
	return func(impl *WorkImpl) {
		impl.Crypto = crypto
	}
}

// ClearTempOption ...
func ClearTempOption(b bool) WorkOptions {
	return func(impl *WorkImpl) {
		impl.ClearTemp = b
	}
}

func dummy(src []byte) (IVideo, error) {
	log.With("src", string(src)).Panic("dummy")
	return nil, nil
}

func defaultWork(options ...WorkOptions) *WorkImpl {
	impl := &WorkImpl{
		ID:         "",
		Status:     WorkWaiting,
		VideoPaths: nil,
		PosterPath: "",
		ThumbPath:  "",
		SamplePath: nil,
		Scale:      fftool.Scale720P,
		Output:     os.TempDir(),
		Skip:       nil,
		ClearTemp:  true,
	}
	for _, opt := range options {
		opt(impl)
	}
	return impl
}

func newWork(wt string, impl *WorkImpl, val []byte) *Work {
	return &Work{
		WorkImpl: impl,
		WorkType: wt,
		Value:    val,
	}
}

// Reset ...
func (w *Work) Reset() error {
	w.WorkImpl.Status = WorkWaiting
	return w.Update()
}

// Status ...
func (w Work) Status() WorkStatus {
	return w.WorkImpl.Status
}

func (w Work) slice(ctx context.Context, input string) (*Fragment, error) {
	format, e := _ffprobe.StreamFormat(input)
	//format, e := split.FFProbeStreamFormat(input)
	if e != nil {
		return nil, Wrap(e, "probe error")
	}
	if !IsMedia(format) {
		return nil, errors.New("file is not a video/audio")
	}
	cfg := fftool.DefaultConfig()
	cfg.SetSlice(true)
	cfg.OutputPath = w.Output()
	cfg.Scale = w.WorkImpl.Scale
	if w.Crypto != nil {
		cfg.SetCrypt(*w.Crypto)
	}

	sharpness := fmt.Sprintf("%dP", fftool.ScaleValue(w.WorkImpl.Scale))
	ff := fftool.NewFFMpeg(cfg)

	ff = ff.OptimizeWithFormat(format)

	e = ff.Run(ctx, input)
	//sa, e := split.FFMpegSplitToM3U8(ctx, input, split.StreamFormatOption(format), split.ScaleOption(formatScale(w.Scale)), split.OutputOption(w.Output()), split.AutoOption(true))
	if e != nil {
		return nil, Wrap(e)
	}

	return &Fragment{
		scale:     cfg.Scale,
		output:    cfg.ProcessPath(),
		skip:      w.Skip,
		input:     input,
		sharpness: sharpness,
	}, nil
}

func (w Work) video() (IVideo, error) {
	fn, b := WorkRunProcessFunction[w.WorkType]
	if !b {
		return nil, errors.New("work type not found")
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

// Output ...
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

// Stop ...
func (w *Work) Stop() error {
	if w.cancel != nil {
		w.cancel()
	}
	w.WorkImpl.Status = WorkStopped
	e := w.Update()
	if e != nil {
		return e
	}
	return nil
}

// CheckStop ...
func (w Work) CheckStop(f func() error) error {
	select {
	case <-w.ctx.Done():
		return w.ctx.Err()
	default:
		return f()
	}
}

// Run ...
func (w *Work) Run(ctx context.Context) (e error) {
	w.ctx, w.cancel = context.WithCancel(ctx)
	defer w.cancel()
	w.WorkImpl.Status = WorkRunning
	if err := w.Update(); err != nil {
		return Wrap(err, "run update")
	}
	v, e := w.video()
	if e != nil {
		return Wrap(e, "run video")
	}
	for _, path := range w.VideoPaths {
		if path == "" {
			continue
		}

		video := v.Video()
		video.TotalEpisode = strconv.Itoa(len(w.VideoPaths))
		video.Episode = strconv.Itoa(GetFileIndex(path))
		if err := w.CheckStop(func() error {
			if !ExistVerifyString("source", w.Skip...) {
				s, e := globalNode.AddFile(ctx, path)
				if e != nil {
					return Wrap(e, "add source")
				}
				video.SourceHash = s
			}
			return nil
		}); err != nil {
			return err
		}

		if err := w.CheckStop(func() error {
			if !ExistVerifyString("slice", w.Skip...) {
				f, e := w.slice(ctx, path)
				if e != nil {
					return Wrap(e, "run slice")
				}
				s, e := globalNode.AddDir(ctx, f.Output())
				if e != nil {
					return Wrap(e, "add slice")
				}
				video.M3U8Hash = s
			}
			return nil
		}); err != nil {
			return err
		}

		if err := w.CheckStop(func() error {
			if !ExistVerifyString("poster", w.Skip...) && w.PosterPath != "" {
				s, e := globalNode.AddFile(ctx, w.PosterPath)
				if e != nil {
					return Wrap(e, "add poster")
				}
				video.PosterHash = s
			}
			return nil
		}); err != nil {
			return err
		}
		if err := w.CheckStop(func() error {
			if !ExistVerifyString("thumb", w.Skip...) && w.ThumbPath != "" {
				s, e := globalNode.AddFile(ctx, w.ThumbPath)
				if e != nil {
					return Wrap(e, "add thumb")
				}
				video.ThumbHash = s
			}
			return nil
		}); err != nil {
			return err
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
	return Wrap(w.Update(), "finished")
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

// Work ...
func (w *Work) Work() *Work {
	wc := *w
	wc.Value = make([]byte, len(w.Value))
	copy(wc.Value, w.Value)
	return &wc
}

// Video ...
func (w *Work) Video() *Video {
	v, err := w.video()
	if err != nil {
		return nil
	}
	return v.Video()
}

// Info ...
func (w *Work) Info() string {
	return string(w.Value)
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
func Wrap(err error, msg ...string) error {
	if err == nil {
		return nil
	}
	if msg != nil {
		return fmt.Errorf("%s:%w", strings.Join(msg, ""), err)
	}
	return fmt.Errorf("%w", err)
}
