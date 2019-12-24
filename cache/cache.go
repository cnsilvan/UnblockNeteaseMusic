package cache

import (
	"UnblockNeteaseMusic/common"
	"sync"
)

var cache sync.Map

func Put(key interface{}, value interface{}) {
	cache.Store(key, value)
}
func Get(key interface{}) (interface{}, bool) {
	return cache.Load(key)
}
func GetSong(key interface{}) (common.Song, bool) {
	var song common.Song
	if value, ok := cache.Load(key); ok {
		song = value.(common.Song)
		return song, ok
	}
	return song, false
}
func Delete(key interface{}) {
	cache.Delete(key)
}
