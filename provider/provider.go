package provider

import (
	"UnblockNeteaseMusic/common"
	"UnblockNeteaseMusic/host"
	"UnblockNeteaseMusic/network"
	"UnblockNeteaseMusic/provider/kuwo"
	"UnblockNeteaseMusic/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var cache = make(map[string]common.Song)

func UpdateCacheMd5(songId string, songMd5 string) {
	if song, ok := cache[songId]; ok {
		song.Md5 = songMd5
		cache[songId] = song
		//fmt.Println("update cache,songId:", songId, ",md5:", songMd5, utils.ToJson(song))
	}
}
func Find(id string) common.Song {
	fmt.Println("find song info,id:", id)
	//if song, ok := cache[id]; ok {
	//	fmt.Println("hit cache:", utils.ToJson(song))
	//	return song
	//}

	var songT common.Song
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
		oJson := utils.ParseJson(body)
		if oJson["songs"] != nil {
			song := oJson["songs"].(common.SliceType)[0]
			var searchSong = make(common.MapType, 6)
			var artists []string
			switch song.(type) {
			case common.MapType:
				searchSong["id"] = song.(common.MapType)["id"]
				searchSong["name"] = song.(common.MapType)["name"]
				searchSong["alias"] = song.(common.MapType)["alias"]
				searchSong["duration"] = song.(common.MapType)["duration"]
				searchSong["album"] = make(common.MapType, 2)
				searchSong["album"].(common.MapType)["id"] = song.(common.MapType)["album"].(common.MapType)["id"]
				searchSong["album"].(common.MapType)["name"] = song.(common.MapType)["album"].(common.MapType)["name"]
				switch song.(common.MapType)["artists"].(type) {
				case common.SliceType:
					length := len(song.(common.MapType)["artists"].(common.SliceType))
					searchSong["artists"] = make(common.SliceType, length)
					artists = make([]string, length)
					for index, value := range song.(common.MapType)["artists"].(common.SliceType) {
						if searchSong["artists"].(common.SliceType)[index] == nil {
							searchSong["artists"].(common.SliceType)[index] = make(common.MapType, 2)
						}
						searchSong["artists"].(common.SliceType)[index].(common.MapType)["id"] = value.(common.MapType)["id"]
						searchSong["artists"].(common.SliceType)[index].(common.MapType)["name"] = value.(common.MapType)["name"]
						artists[index] = value.(common.MapType)["name"].(string)
					}

				}
			default:

			}
			if searchSong["name"] != nil {
				searchSong["name"] = utils.ReplaceAll(searchSong["name"].(string), `\s*cover[:：\s][^）]+）`, "")
				searchSong["name"] = utils.ReplaceAll(searchSong["name"].(string), `\(\s*cover[:：\s][^\)]+\)`, "")
			}
			searchSong["keyword"] = searchSong["name"].(string) + " " + strings.Join(artists, " / ")
			songUrl := searchSongFn(searchSong)
			if len(songUrl.Url) > 0 { //未版权
				songS := processSong(songUrl)
				if songS.Size > 0 {
					//fmt.Println(utils.ToJson(songS))
					if s, ok := cache[id]; ok && len(s.Md5) > 0 {
						songS.Md5 = s.Md5
					}
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
func searchSongFn(key common.MapType) common.Song {
	//cache after
	//kugou.SearchSong(key)
	return kuwo.SearchSong(key)

}
func processSong(song common.Song) common.Song {
	if len(song.Url) > 0 {
		if len(song.Md5) > 0 && song.Br > 0 && song.Size > 0 {
			return song
		}
		if song.Br > 0 && song.Size > 0 && !strings.Contains(song.Url, "qq.com") && !strings.Contains(song.Url, "xiami.net") && !strings.Contains(song.Url, "qianqian.com") {
			return song
		}
		header := make(http.Header, 1)
		header["range"] = append(header["range"], "bytes=0-8191")
		clientRequest := network.ClientRequest{
			Method:    http.MethodGet,
			RemoteUrl: song.Url,
			Header:    header,
			Proxy:     false,
		}
		resp, err := network.Request(&clientRequest)
		if err != nil {
			fmt.Println("processSong fail:", err)
			return song
		}
		if resp.StatusCode > 199 && resp.StatusCode < 300 {
			if strings.Contains(song.Url, "qq.com") {
				song.Md5 = resp.Header.Get("server-md5")
			} else if strings.Contains(song.Url, "xiami.net") || strings.Contains(song.Url, "qianqian.com") {
				song.Md5 = strings.ToLower(utils.ReplaceAll(resp.Header.Get("etag"), `/"/g`, ""))
				//.replace(/"/g, '').toLowerCase()
			}
			if song.Size == 0 {
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
			}
			if song.Br == 0 {
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
			}
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
