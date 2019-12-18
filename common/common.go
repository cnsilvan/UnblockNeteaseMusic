package common

type MapType = map[string]interface{}
type SliceType = []interface{}
type Song struct {
	Size int64
	Br   int
	Url  string
	Md5  string
}
type SearchSong struct {
	Name   string
	Artist string
	Url    string
}
