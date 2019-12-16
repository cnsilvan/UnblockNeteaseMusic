package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"host"
	"net/http"
	"network"
	"provider/kuwo"
	"strconv"
	"strings"
	"utils"
)

type Song struct {
	Size int64
	Br   int
	Url  string
	Md5  string
}
type MapType = map[string]interface{}
type SliceType = []interface{}

var cache = make(map[string]Song)

func UpdateCacheMd5(songId string, songMd5 string) {
	if song, ok := cache[songId]; ok {
		song.Md5 = songMd5
		cache[songId] = song
		//fmt.Println("update cache,songId:", songId, ",md5:", songMd5, utils.ToJson(song))
	}
}
func Find(id string) Song {
	fmt.Println("find song info,id:", id)
	if song, ok := cache[id]; ok {
		fmt.Println("hit cache:", utils.ToJson(song))
		return song
	}

	var songT Song
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: "https://" + host.ProxyDomain["music.163.com"] + "/api/song/detail?ids=[" + id + "]",
		Host:      "music.163.com",
		Header:    nil,
		Proxy:     true,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		return songT
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err2 := network.GetResponseBody(resp, false)
		if err2 != nil {
			fmt.Println("GetResponseBody fail")
			return songT
		}
		var oJson MapType
		d := json.NewDecoder(bytes.NewReader(body))
		d.UseNumber()
		d.Decode(&oJson)
		if oJson["songs"] != nil {
			song := oJson["songs"].(SliceType)[0]
			var modifiedJson = make(MapType, 6)
			var artists []string
			switch song.(type) {
			case MapType:
				modifiedJson["id"] = song.(MapType)["id"]
				modifiedJson["name"] = song.(MapType)["name"]
				modifiedJson["alias"] = song.(MapType)["alias"]
				modifiedJson["duration"] = song.(MapType)["duration"]
				modifiedJson["album"] = make(MapType, 2)
				modifiedJson["album"].(MapType)["id"] = song.(MapType)["album"].(MapType)["id"]
				modifiedJson["album"].(MapType)["name"] = song.(MapType)["album"].(MapType)["name"]
				switch song.(MapType)["artists"].(type) {
				case SliceType:
					length := len(song.(MapType)["artists"].(SliceType))
					modifiedJson["artists"] = make(SliceType, length)
					artists = make([]string, length)
					for index, value := range song.(MapType)["artists"].(SliceType) {
						if modifiedJson["artists"].(SliceType)[index] == nil {
							modifiedJson["artists"].(SliceType)[index] = make(MapType, 2)
						}
						modifiedJson["artists"].(SliceType)[index].(MapType)["id"] = value.(MapType)["id"]
						modifiedJson["artists"].(SliceType)[index].(MapType)["name"] = value.(MapType)["name"]
						artists[index] = value.(MapType)["name"].(string)
					}

				}
			default:

			}
			if modifiedJson["name"] != nil {
				modifiedJson["name"] = utils.ReplaceAll(modifiedJson["name"].(string), `\s*cover[:：\s][^）]+）`, "")
				modifiedJson["name"] = utils.ReplaceAll(modifiedJson["name"].(string), `\(\s*cover[:：\s][^\)]+\)`, "")
			}
			modifiedJson["keyword"] = modifiedJson["name"].(string) + " - " + strings.Join(artists, " / ")
			songUrl := searchSong(modifiedJson)
			if len(songUrl) > 0 { //未版权
				songS := processSong(songUrl)
				if songS.Size > 0 {
					//fmt.Println(utils.ToJson(songS))
					cache[id] = songS
					return songS
				}
			}
			//fmt.Println(utils.ToJson(modifiedJson))
			return songT
		} else {
			return songT
		}
	} else {
		return songT
	}

}
func searchSong(key MapType) string {
	//cache after
	return kuwo.SearchSong(key)

}
func processSong(songUrl string) Song {
	var song Song
	if len(songUrl) > 0 {
		header := make(http.Header, 1)
		header["range"] = append(header["range"], "bytes=0-8191")
		clientRequest := network.ClientRequest{
			Method:    http.MethodGet,
			RemoteUrl: songUrl,
			Header:    header,
			Proxy:     false,
		}
		resp, err := network.Request(&clientRequest)
		if err != nil {
			fmt.Println("processSong fail:", err)
			return song
		}
		if resp.StatusCode > 199 && resp.StatusCode < 300 {
			if strings.Contains(songUrl, "qq.com") {
				song.Md5 = resp.Header.Get("server-md5")
			} else if strings.Contains(songUrl, "xiami.net") || strings.Contains(songUrl, "qianqian.com") {
				song.Md5 = strings.ToLower(utils.ReplaceAll(resp.Header.Get("etag"), `/"/g`, ""))
				//.replace(/"/g, '').toLowerCase()
			}
			size := resp.Header.Get("content-range")
			if len(size) > 0 {
				sizeSlice := strings.Split(size, "/")
				if len(sizeSlice) > 0 {
					size = sizeSlice[len(sizeSlice)-1]
				}
			} else {
				size = resp.Header.Get("content-length")
				if len(size) < 1 {
					size = "0"
				}
			}
			song.Size, _ = strconv.ParseInt(size, 10, 64)
			song.Url = songUrl
			if resp.Header.Get("content-length") == "8192" {
				body, err := network.GetResponseBody(resp, false)
				if err != nil {
					fmt.Println("song GetResponseBody error:", err)
					return song
				}
				bitrate := decodeBitrate(body)
				if bitrate > 0 && bitrate < 500 {
					song.Br = bitrate * 1000
				}
			}
			//song.url = response.url.href
		}
	}
	return song
}
func decodeBitrate(data []byte) int {
	bitRateMap := map[int]map[int][]int{
		0: {
			3: {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 500},
			2: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
			1: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
		},
		3: {
			3: {0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 500},
			2: {0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, 500},
			1: {0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 500},
		},
		2: {
			3: {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 500},
			2: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
			1: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 500},
		},
	}

	var pointer = 0
	if strings.EqualFold(string(data[0:4]), "fLaC") {
		return 999
	}
	if strings.EqualFold(string(data[0:3]), "ID3") {
		pointer = 6
		var size = 0
		for index, value := range data[pointer : pointer+4] {
			size = size + int((value&0x7f)<<(7*(3-index)))
		}
		pointer = 10 + size
	}

	header := data[pointer : pointer+4]
	// https://www.allegro.cc/forums/thread/591512/674023
	if len(header) == 4 &&
		header[0] == 0xff &&
		((header[1]>>5)&0x7) == 0x7 &&
		((header[1]>>1)&0x3) != 0 &&
		((header[2]>>4)&0xf) != 0xf &&
		((header[2]>>2)&0x3) != 0x3 {
		version := (header[1] >> 3) & 0x3
		layer := (header[1] >> 1) & 0x3
		bitrate := header[2] >> 4
		return bitRateMap[int(version)][int(layer)][int(bitrate)]
	}
	return 0
}
