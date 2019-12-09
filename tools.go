package conversion

import (
	"github.com/glvd/go-fftool"
)

type Scale = fftool.Scale
type Config = fftool.Config

var SliceConfig = fftool.DefaultConfig()

//var _ffmpeg *fftool.FFMpeg
var _ffprobe *fftool.FFProbe

func InitFFTool() {
	//_ffmpeg = fftool.NewFFMpeg(SliceConfig)
	_ffprobe = fftool.NewFFProbe()
}
