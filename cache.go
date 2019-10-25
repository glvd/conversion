package conversion

import (
	"github.com/gocacher/badger-cache"
	"github.com/gocacher/cacher"
)

var CachePath = cache.DefaultPath

func SetCachePath(path string) {
	CachePath = path
}

func RegisterCache() {
	cacher.Register(cache.NewBadgerCache(CachePath))
}
