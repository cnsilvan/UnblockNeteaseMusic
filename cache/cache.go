package cache

import (
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"sync"
)

var cache sync.Map

//func Put(key string, value interface{}) {
//	cache.Store(key, value)
//}
//func Get(key interface{}) (interface{}, bool) {
//	return cache.Load(key)
//}
func PutSong(key common.SearchMusic, value common.Song) {
	cache.Store(fmt.Sprintf("%+v", key), value)
}
func GetSong(key common.SearchMusic) (common.Song, bool) {
	var song common.Song
	if value, ok := cache.Load(fmt.Sprintf("%+v", key)); ok {
		song = value.(common.Song)
		return song, ok
	}
	return song, false
}
func Delete(key common.SearchMusic) {
	cache.Delete(fmt.Sprintf("%+v", key))
}
