package conversion

import (
	cache "github.com/gocacher/badger-cache"
	"github.com/gocacher/cacher"
)

var _cache cacher.Cacher

// CachePath ...
var CachePath = cache.DefaultPath

// SetCachePath ...
func SetCachePath(path string) {
	CachePath = path
}

// RegisterCache ...
func RegisterCache() {
	_cache = cache.NewBadgerCache(CachePath)
	cacher.Register(_cache)
}
