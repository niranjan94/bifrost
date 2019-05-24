package config

import (
	"github.com/niranjan94/bifrost/utils"
	c "github.com/patrickmn/go-cache"
	"time"
)

var cache *c.Cache

func init() {
	cache = c.New(5*time.Minute, 10*time.Minute)
}

// ToCache stores a value to cache at a key
func ToCache(key string, value interface{}) {
	utils.Must(cache.Add(key, value, c.NoExpiration))
}

// FromCache returns the cached value for a key
func FromCache(key string) (interface{}, bool) {
	return cache.Get(key)
}

// MemoizedFn stores the result of the function in the cache and returns it on subsequent calls
func MemoizedFn(key string, keyPrefix string, fn func(string) interface{}) interface{} {
	cacheKey := keyPrefix + ":" + key
	if cachedValue, found := FromCache(cacheKey); found {
		return *cachedValue.(*interface{})
	}
	value := fn(key)
	if stringValue, isString := value.(string); isString {
		if isReference(stringValue) {
			value = MemoizedFn(stringValue, keyPrefix, fn)
		}
	}
	ToCache(cacheKey, &value)
	return value
}