package conversion

import (
	"github.com/glvd/go-fftool"
)

// Scale ...
type Scale = fftool.Scale

// Config ...
type Config = fftool.Config

var _ffprobe *fftool.FFProbe

// InitFFTool ...
func InitFFTool() {
	_ffprobe = fftool.NewFFProbe()
}
