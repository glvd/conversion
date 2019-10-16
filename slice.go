package conversion

import (
	"errors"
	"fmt"
	"os"

	cmd "github.com/godcong/go-ffmpeg-cmd"
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

// Slice ...
type Slice struct {
	Scale       Scale
	SliceOutput string
	SkipType    []interface{}
	SkipExist   bool
	SkipSlice   bool
	file        string
}

// NewSlice ...
func NewSlice() *Slice {
	output := os.TempDir()
	return &Slice{
		SliceOutput: output,
	}
}

// SliceCallbackFunc ...
type SliceCallbackFunc func(s *Slice, sa *cmd.SplitArgs, v interface{}) (e error)

func scale(scale Scale) int {
	switch scale {
	case 480, 1080:
		return int(scale)
	default:
		return 720
	}
}

func scaleStr(s Scale) string {
	return fmt.Sprintf("%dP", scale(s))
}

func isMedia(format *cmd.StreamFormat) bool {
	video := format.Video()
	audio := format.Audio()
	if audio == nil || video == nil {
		return false
	}
	return true
}

// Call ...
func (c *sliceCall) Call(s *Slice) (e error) {
	sa, e := sliceVideo(s, c.file, c.hash)
	if e != nil {
		return e
	}
	return c.cb(s, sa, c.hash)
}

func sliceVideo(slice *Slice, file string, u *Hash) (sa *cmd.SplitArgs, e error) {
	format, e := cmd.FFProbeStreamFormat(file)
	if e != nil {
		return nil, e
	}
	if !isMedia(format) {
		return nil, errors.New("file is not a video/audio")
	}

	//u.Type = model.TypeSlice
	s := slice.Scale
	if s != 0 {
		res := format.ResolutionInt()
		if int64(res) < int64(s) {
			s = Scale(res)
		}
		sa, e = cmd.FFMpegSplitToM3U8(nil, file, cmd.StreamFormatOption(format), cmd.ScaleOption(int64(s)), cmd.OutputOption(slice.SliceOutput))
		u.Sharpness = scaleStr(s)
	} else {
		sa, e = cmd.FFMpegSplitToM3U8(nil, file, cmd.StreamFormatOption(format), cmd.OutputOption(slice.SliceOutput))
		u.Sharpness = format.Resolution() + "P"
	}

	log.Infof("%+v", sa)
	return
}

type sliceCall struct {
	cb   SliceCallbackFunc
	file string
	hash *Hash
}

var _ SliceCaller = &sliceCall{}
