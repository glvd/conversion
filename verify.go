package conversion

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/glvd/split"
)

// IsPicture ...
func IsPicture(name string) bool {
	picture := ".bmp,.jpg,.png,.tif,.gif,.pcx,.tga,.exif,.fpx,.svg,.psd,.cdr,.pcd,.dxf,.ufo,.eps,.ai,.raw,.wmf,.webp"
	ext := filepath.Ext(name)
	return strings.Index(picture, ext) != -1
}

// IsVideo ...
func IsVideo(filename string) bool {
	video := `.swf,.flv,.3gp,.ogm,.vob,.m4v,.mkv,.mp4,.mpg,.mpeg,.avi,.rm,.rmvb,.mov,.wmv,.asf,.dat,.asx,.wvx,.mpe,.mpa`
	ext := path.Ext(filename)
	return strings.Index(video, ext) != -1
}

// IsMedia ...
func IsMedia(format *split.StreamFormat) bool {
	video := format.Video()
	audio := format.Audio()
	if audio == nil || video == nil {
		return false
	}
	return true
}

// ExistVerifyString ...
func ExistVerifyString(tp string, v ...string) bool {
	for i := range v {
		if v[i] == tp {
			return true
		}
	}
	return false
}

// ExistVerify ...
func ExistVerifyFunc(tp string, f func(interface{}) string, v ...interface{}) bool {
	for i := range v {
		if f(v[i]) == tp {
			return true
		}
	}
	return false
}
