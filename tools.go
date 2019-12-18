package conversion

import (
	"github.com/glvd/go-fftool"
)

// Scale ...
type Scale = fftool.Scale

// Crypto ...
type Crypto = fftool.Crypto

// Config ...
type Config = fftool.Config

var _ffprobe *fftool.FFProbe

// InitFFTool ...
func InitFFTool() {
	_ffprobe = fftool.NewFFProbe()
}
