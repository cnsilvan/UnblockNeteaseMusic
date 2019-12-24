package common

import (
	"math/rand"
	"time"
)

type MapType = map[string]interface{}
type SliceType = []interface{}
type Song struct {
	Id         string
	Size       int64
	Br         int
	Url        string
	Md5        string
	Name       string
	Artist     string
	MatchScore float32
}

var (
	ProxyIp     = "127.0.0.1"
	ProxyDomain = map[string]string{
		"music.163.com":            "59.111.181.35",
		"interface.music.163.com":  "59.111.181.35",
		"interface3.music.163.com": "59.111.181.35",
		"apm.music.163.com":        "59.111.181.35",
		"apm3.music.163.com":       "59.111.181.35",
	}
	HostDomain = map[string]string{
		"music.163.com":           "59.111.181.35",
		"interface.music.163.com": "59.111.181.35",
	}
	Source []string
	Rand   = rand.New(
		rand.NewSource(time.Now().UnixNano()))
)
