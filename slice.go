package conversion

import (
	"errors"
	"os"
	"strconv"

	"github.com/glvd/split"
)

// Scale ...
type Scale int64

// HighScale ...
const HighScale Scale = 1080

// MiddleScale ...
const MiddleScale Scale = 720

// LowScale ...
const LowScale Scale = 480

// SliceCaller ...
type SliceCaller interface {
	Call(*Slice) error
}

// SliceOptions ...
type SliceOptions func(slice *Slice)

// Slice ...
type Slice struct {
	scale     Scale
	output    string
	skip      []interface{}
	skipExist bool
	input     string
}

// SliceSkip ...
func SliceSkip(skip ...interface{}) SliceOptions {
	return func(slice *Slice) {
		slice.skip = skip
	}
}

// SliceOutput ...
func SliceOutput(output string) SliceOptions {
	return func(slice *Slice) {
		slice.output = output
	}
}

// SliceScale ...
func SliceScale(scale Scale) SliceOptions {
	return func(slice *Slice) {
		slice.scale = scale
	}
}

// NewSlice ...
func NewSlice(input string, options ...SliceOptions) *Slice {
	output := os.TempDir()
	slice := &Slice{
		output: output,
		scale:  MiddleScale,
		input:  input,
	}
	for _, opts := range options {
		opts(slice)
	}
	return slice
}

// Scale ...
func (s Slice) Scale() int64 {
	return scale(s.scale)
}

func toScale(scale int64) Scale {
	if scale > 1080 {
		return HighScale
	} else if scale > 720 {
		return MiddleScale
	}
	return LowScale
}

func scale(scale Scale) int64 {
	switch scale {
	case 480, 1080:
		return int64(scale)
	default:
		return 720
	}
}

func isMedia(format *split.StreamFormat) bool {
	video := format.Video()
	audio := format.Audio()
	if audio == nil || video == nil {
		return false
	}
	return true
}

func sliceVideo(slice *Slice, file string, hash *Hash) (sa *split.Argument, e error) {
	format, e := split.FFProbeStreamFormat(file)
	if e != nil {
		return nil, e
	}
	if !isMedia(format) {
		return nil, errors.New("file is not a video/audio")
	}

	if SkipVerify("slice", slice.skip...) {
		res := toScale(int64(format.ResolutionInt()))
		if res < slice.scale {
			slice.scale = res
		}
		sa, e = split.FFMpegSplitToM3U8(nil, file, split.StreamFormatOption(format), split.ScaleOption(slice.Scale()), split.OutputOption(slice.output))
		hash.Sharpness = strconv.FormatInt(slice.Scale(), 10) + "P"
		log.Infof("%+v", sa)
		return sa, e
	}
	return nil, errors.New("slice skipped")
}
