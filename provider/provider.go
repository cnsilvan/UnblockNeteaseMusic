package provider

import (
	"UnblockNeteaseMusic/cache"
	"UnblockNeteaseMusic/common"
	"UnblockNeteaseMusic/network"
	kugou "UnblockNeteaseMusic/provider/kugou"
	"UnblockNeteaseMusic/provider/kuwo"
	"UnblockNeteaseMusic/provider/migu"
	"UnblockNeteaseMusic/utils"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func UpdateCacheMd5(songId string, songMd5 string) {
	if song, ok := cache.GetSong(songId); ok {
		song.Md5 = songMd5
		cache.Put(songId, song)
	}
}
func Find(id string) common.Song {
	fmt.Println("find song info,id:", id)
	if song, ok := cache.GetSong(id); ok {
		fmt.Println("hit cache:", utils.ToJson(song))
		if checkCache(song) {
			return song
		}else{
			fmt.Println("but cache invalid")
		}
	}
	var songT common.Song
	songT.Id = id
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: "https://" + common.HostDomain["music.163.com"] + "/api/song/detail?ids=[" + id + "]",
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
		body, err2 := network.StealResponseBody(resp)
		if err2 != nil {
			fmt.Println("GetResponseBody fail")
			return songT
		}
		oJson := utils.ParseJsonV2(body)
		if songs, ok := oJson["songs"].(common.SliceType); ok && len(songs) > 0 {
			song := songs[0]
			var searchSong = make(common.MapType, 8)
			searchSong["songId"] = id
			var artists []string
			switch song.(type) {
			case common.MapType:
				neteaseSong := song.(common.MapType)
				searchSong["id"] = neteaseSong["id"]
				searchSong["name"] = neteaseSong["name"]
				searchSong["alias"] = neteaseSong["alias"]
				searchSong["duration"] = neteaseSong["duration"]
				searchSong["album"] = make(common.MapType, 2)
				searchSong["album"].(common.MapType)["id"], ok = neteaseSong["album"].(common.MapType)["id"]
				searchSong["album"].(common.MapType)["name"], ok = neteaseSong["album"].(common.MapType)["name"]
				switch neteaseSong["artists"].(type) {
				case common.SliceType:
					length := len(neteaseSong["artists"].(common.SliceType))
					searchSong["artists"] = make(common.SliceType, length)
					artists = make([]string, length)
					for index, value := range neteaseSong["artists"].(common.SliceType) {
						if searchSong["artists"].(common.SliceType)[index] == nil {
							searchSong["artists"].(common.SliceType)[index] = make(common.MapType, 2)
						}
						searchSong["artists"].(common.SliceType)[index].(common.MapType)["id"], ok = value.(common.MapType)["id"]
						searchSong["artists"].(common.SliceType)[index].(common.MapType)["name"], ok = value.(common.MapType)["name"]
						artists[index], ok = value.(common.MapType)["name"].(string)
					}

				}
			default:

			}
			//if searchSong["name"] != nil {
			//	searchSong["name"] = utils.ReplaceAll(searchSong["name"].(string), `\s*cover[:：\s][^）]+）`, "")
			//	searchSong["name"] = utils.ReplaceAll(searchSong["name"].(string), `(\s*cover[:：\s][^\)]+)`, "")
			//}
			searchSong["artistsName"] = strings.Join(artists, " ")
			searchSong["keyword"] = searchSong["name"].(string) + " " + searchSong["artistsName"].(string)
			fmt.Println("search song:" + searchSong["keyword"].(string))
			songT = searchSongFn(searchSong)
			fmt.Println(utils.ToJson(songT))
			return songT

		} else {
			return songT
		}
	} else {
		return songT
	}
}

func searchSongFn(key common.MapType) common.Song {
	id := "0"
	searchSongName := key["name"].(string)
	searchSongName = strings.ToUpper(searchSongName)
	searchArtistsName := key["artistsName"].(string)
	searchArtistsName = strings.ToUpper(searchArtistsName)
	if songId, ok := key["songId"]; ok {
		id = songId.(string)
	}
	var ch = make(chan common.Song)
	now := time.Now().UnixNano() / 1e6
	songs := getSongFromAllSource(key, ch)
	//fmt.Println(utils.ToJson(songs))
	fmt.Println("consumed:", (time.Now().UnixNano()/1e6)-now, "ms")
	result := common.Song{}
	result.Size = 0
	for _, song := range songs {
		//songNameKeys := utils.ParseSongNameKeyWord(song.Name)
		//fmt.Println("songNameKeys:", strings.Join(songNameKeys, "、"))
		//songNameSores := utils.CalMatchScores(searchSongName, songNameKeys)
		//fmt.Println("songNameSores:", songNameSores)
		//artistKeys := utils.ParseSingerKeyWord(song.Artist)
		//fmt.Println("artistKeys:", strings.Join(artistKeys, "、"))
		//artistsNameSores := utils.CalMatchScores(searchArtistsName, artistKeys)
		//fmt.Println("artistsNameSores:", artistsNameSores)
		//songMatchScore := songNameSores*0.6 + artistsNameSores*0.4
		//song.MatchScore = songMatchScore
		if song.MatchScore > result.MatchScore {
			result = song
		} else if song.MatchScore == result.MatchScore && song.Size > result.Size {
			result = song
		}
	}

	if id != "0" {
		result.Id = id
		if len(result.Url) > 0 {
			cache.Put(id, result)
		}
	}
	return result

}

func getSongFromAllSource(key common.MapType, ch chan common.Song) []common.Song {
	var songs []common.Song
	sum := 0
	for _, source := range common.Source {
		switch source {
		case "kuwo":
			go getSongFromKuWo(key, ch)
			sum++
		case "kugou":
			go getSongFromKuGou(key, ch)
			sum++
		case "migu":
			go getSongFromMiGu(key, ch)
			sum++

		}
	}
	for {
		select {
		case song, _ := <-ch:
			if len(song.Url) > 0 {
				songs = append(songs, song)
			}
			sum--
			if sum <= 0 {
				return songs
			}
		case <-time.After(time.Second * 6):
			return songs
		}
	}
}
func getSongFromKuWo(key common.MapType, ch chan common.Song) {
	ch <- calculateSongInfo(kuwo.SearchSong(key))
}
func getSongFromKuGou(key common.MapType, ch chan common.Song) {
	ch <- calculateSongInfo(kugou.SearchSong(key))
}
func getSongFromMiGu(key common.MapType, ch chan common.Song) {
	ch <- calculateSongInfo(migu.SearchSong(key))
}
func calculateSongInfo(song common.Song) common.Song {
	if len(song.Url) > 0 {
		if len(song.Md5) > 0 && song.Br > 0 && song.Size > 0 {
			return song
		}
		if song.Br > 0 && song.Size > 0 && !strings.Contains(song.Url, "qq.com") && !strings.Contains(song.Url, "xiami.net") && !strings.Contains(song.Url, "qianqian.com") {
			return song
		}
		header := make(http.Header, 1)
		header["range"] = append(header["range"], "bytes=0-8191")
		uri, err := url.Parse(song.Url)
		if err == nil {
			song.Url = uri.String()
		}
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
		defer resp.Body.Close()
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
					if bitrate == 999 || (bitrate > 0 && bitrate < 500) {
						song.Br = bitrate * 1000
					}
				}
			}
		} else {
			return common.Song{}
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
func checkCache(song common.Song) bool {
	header := make(http.Header, 1)
	header["range"] = append(header["range"], "bytes=0-1")
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: song.Url,
		Header:    header,
		Proxy:     false,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		fmt.Println("checkCache fail:", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		cache.Delete(song.Id)
		return false
	} else {
		return true
	}
}
