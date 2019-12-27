package kuwo

import (
	"UnblockNeteaseMusic/common"
	"UnblockNeteaseMusic/network"
	"UnblockNeteaseMusic/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func SearchSong(key common.MapType) common.Song {
	searchSong := common.Song{

	}
	keyword := key["keyword"].(string)
	searchSongName := key["name"].(string)
	searchSongName = strings.ToUpper(searchSongName)
	searchArtistsName := key["artistsName"].(string)
	searchArtistsName = strings.ToUpper(searchArtistsName)
	cookies := getCookies()
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: "http://songsearch.kugou.com/song_search_v2?keyword=" + keyword + "&page=1",
		Host:      "songsearch.kugou.com",
		Cookies:   cookies,
		Proxy:     false,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	defer resp.Body.Close()
	//header := resp.Header
	body, err := network.StealResponseBody(resp)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	result := utils.ParseJsonV2(body)
	//fmt.Println(utils.ToJson(result))
	data := result["data"]
	var songHashId = ""
	if data != nil && data.(common.MapType)["lists"] != nil {
		switch data.(type) {
		case common.MapType:
			lists := data.(common.MapType)["lists"]
			switch lists.(type) {
			case common.SliceType:
				listLength := len(lists.(common.SliceType))
				if listLength > 0 {
					for index, matched := range lists.(common.SliceType) {
						if kugouSong, ok := matched.(common.MapType); ok {
							if fileHash, ok := kugouSong["FileHash"].(string); ok {
								singerName, singerNameOk := kugouSong["SingerName"].(string)
								songName, songNameOk := kugouSong["SongName"].(string)
								var songNameSores float32 = 0.0
								if songNameOk {
									songNameKeys := utils.ParseSongNameKeyWord(songName)
									//fmt.Println("songNameKeys:", strings.Join(songNameKeys, "、"))
									songNameSores = utils.CalMatchScores(searchSongName, songNameKeys)
									//fmt.Println("songNameSores:", songNameSores)
								}
								var artistsNameSores float32 = 0.0
								if singerNameOk {
									artistKeys := utils.ParseSingerKeyWord(singerName)
									//fmt.Println("kugou:artistKeys:", strings.Join(artistKeys, "、"))
									artistsNameSores = utils.CalMatchScores(searchArtistsName, artistKeys)
									//fmt.Println("kugou:artistsNameSores:", artistsNameSores)
								}
								songMatchScore := songNameSores*0.6 + artistsNameSores*0.4
								//fmt.Println("kugou:songMatchScore:", songMatchScore)
								if songMatchScore > searchSong.MatchScore {
									searchSong.MatchScore = songMatchScore
									songHashId = fileHash
									searchSong.Name = songName
									searchSong.Artist = singerName
									searchSong.Artist = strings.ReplaceAll(singerName, " ", "")
								}

							}

						}
						if index >= listLength/2 || index > 9 {
							break
						}
					}
				}

			}

		}

	}
	if len(songHashId) > 0 {
		clientRequest := network.ClientRequest{
			Method:               http.MethodGet,
			RemoteUrl:            "http://www.kugou.com/yy/index.php?r=play/getdata&hash=" + songHashId,
			Host:                 "www.kugou.com",
			Cookies:              cookies,
			Header:               nil,
			ForbiddenEncodeQuery: true,
			Proxy:                false,
		}
		//fmt.Println(clientRequest.RemoteUrl)
		resp, err := network.Request(&clientRequest)
		if err != nil {
			fmt.Println(err)
			return searchSong
		}
		defer resp.Body.Close()
		body, err = network.StealResponseBody(resp)
		songData := utils.ParseJsonV2(body)
		data = songData["data"]
		//fmt.Println(utils.ToJson(songData))
		switch data.(type) {
		case common.MapType:
			if data.(common.MapType)["play_url"] != nil {
				songUrl, ok := data.(common.MapType)["play_url"].(string)
				if ok && strings.Index(songUrl, "http") == 0 {
					searchSong.Url = songUrl
					//searchSong.Size, _ = data.(common.MapType)["filesize"].(json.Number).Int64()
					if br, ok := data.(common.MapType)["bitrate"]; ok {
						switch br.(type) {
						case json.Number:
							searchSong.Br, _ = strconv.Atoi(br.(json.Number).String())
						case int:
							searchSong.Br = br.(int)

						}
					}
					searchSong.Br = searchSong.Br * 1000
					return searchSong
				}
			}
		default:
			return searchSong
		}
		//fmt.Println(utils.ToJson(data))
	}
	return searchSong

}
func getCookies() []*http.Cookie {
	cookies := make([]*http.Cookie, 1)
	cookie := &http.Cookie{Name: "kg_mid", Value: createGuid(), Path: "kugou.com", Domain: "kugou.com"}
	cookies[0] = cookie
	return cookies
}
func createGuid() string {
	guid := s4() + s4() + "-" + s4() + "-" + s4() + "-" + s4() + "-" + s4() + s4() + s4()
	return utils.MD5(bytes.NewBufferString(guid).Bytes())
}
func s4() string {
	num := uint64((1 + common.Rand.Float64()) * 0x10000)
	num = num | 0
	return strconv.FormatUint(num, 16)[1:]
}
