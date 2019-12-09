package conversion

import (
	"github.com/glvd/go-fftool"
)

type Scale = fftool.Scale
type Config = fftool.Config

var _ffprobe *fftool.FFProbe

func InitFFTool() {
	_ffprobe = fftool.NewFFProbe()
}
