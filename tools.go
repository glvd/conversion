package conversion

import (
	"github.com/glvd/go-fftool"
)

var _ffcfg = fftool.DefaultConfig()
var _ffmpeg *fftool.FFMpeg
var _ffprobe *fftool.FFProbe

func init() {
	_ffmpeg = fftool.NewFFMpeg(_ffcfg)
	_ffprobe = fftool.NewFFProbe()
}
