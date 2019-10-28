package conversion

import (
	"github.com/gocacher/badger-cache"
	"github.com/gocacher/cacher"
)

// CachePath ...
var CachePath = cache.DefaultPath

// SetCachePath ...
func SetCachePath(path string) {
	CachePath = path
}

// RegisterCache ...
func RegisterCache() {
	cacher.Register(cache.NewBadgerCache(CachePath))
}
