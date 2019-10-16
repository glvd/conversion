package conversion

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	format *split.StreamFormat

	scale     Scale
	output    string
	skip      []string
	input     string
	sharpness string
}

// SliceSkip ...
func SliceSkip(skip ...string) SliceOptions {
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

// Sharpness ...
func (s Slice) Sharpness() string {
	return s.sharpness
}

// Do ...
func (s *Slice) Do(ctx context.Context) (e error) {
	s.format, e = split.FFProbeStreamFormat(s.input)
	if e != nil {
		return e
	}
	if !isMedia(s.format) {
		return errors.New("file is not a video/audio")
	}
	res := toScale(int64(s.format.ResolutionInt()))
	if res < s.scale {
		s.scale = res
	}
	s.sharpness = strconv.FormatInt(s.Scale(), 10) + "P"
	s.output = filepath.Join(s.output, UUID().String())
	if SkipVerifyString("slice", s.skip...) {
		sa, e := split.FFMpegSplitToM3U8(ctx, s.input, split.StreamFormatOption(s.format), split.ScaleOption(s.Scale()), split.OutputOption(s.output), split.AutoOption(false))
		if e != nil {
			return fmt.Errorf("%w", e)
		}
		log.Infof("%+v", sa)
		return nil
	}
	return errors.New("slice skipped")
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
