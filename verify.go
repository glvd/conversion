package seed

import (
	"github.com/glvd/split"
	"path"
	"path/filepath"
	"strings"
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

func IsMedia(format *split.StreamFormat) bool {
	video := format.Video()
	audio := format.Audio()
	if audio == nil || video == nil {
		return false
	}
	return true
}

// SkipVerifyString ...
func SkipVerifyString(tp string, v ...string) bool {
	for i := range v {
		if v[i] == tp {
			return true
		}
	}
	return false
}

// SkipVerify ...
func SkipVerify(tp string, v ...interface{}) bool {
	for i := range v {
		if v1, b := (v[i]).(string); b {
			if v1 == tp {
				return true
			}
		}
	}
	return false
}
