package cache

import "github.com/muesli/cache2go"

var Cache *cache2go.CacheTable

func NewCache(name string) {
	Cache = cache2go.Cache(name)
}
