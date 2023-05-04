// cache.go: managing cache.

package cache

import (
	"dollwipe/engine"
	"dollwipe/network"
	"os"
	"strings"

	"github.com/neuroliptica/logger"
)

var CacheLogger = logger.MakeLogger("cache").BindToDefault()

// Split provided path into two parts: filename and path to it's directory.
func GetPath(dir string) (string, string) {
	parts := strings.Split(dir, "/")
	return parts[len(parts)-1],
		strings.Join(parts[:len(parts)-1], "/")
}

// General interface for types, which instances can be cached in some format.
type Cacheable interface {
	CachePack(string) error
	CacheUnpack(string) (Cacheable, error)
}

type PostsCache map[network.Proxy]*engine.Post

func (packed PostsCache) CachePack(dir string) error {
	posts := map[network.Proxy]*engine.Post(packed)
	cache := make([]string, 0)
	for key := range posts {
		if key.NoProxy() {
			continue
		}
		proxy := key.Protocol + "://"
		if key.NeedAuth() {
			proxy += key.Login + ":" + key.Pass + "@"
		}
		proxy += key.String()
		cache = append(cache, proxy)
	}
	if len(cache) == 0 {
		return nil
	}
	fname, path := GetPath(dir)
	if fname == "" {
		fname = "cache_tmp"
		CacheLogger.Logf("using default name => ./%s", fname)
		dir = path + "/" + fname
	}
	var err error
	if path != "" {
		err = os.MkdirAll(path, 0750)
		if err != nil {
			return err
		}
	}
	err = os.WriteFile(dir, []byte(strings.Join(cache, "\n")), 0660)
	return err
}
