package kuwo

import (
	"UnblockNeteaseMusic/common"
	"UnblockNeteaseMusic/network"
	"UnblockNeteaseMusic/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func SearchSong(key common.MapType) common.Song {
	searchSong := common.Song{

	}
	keyword := key["keyword"].(string)
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
	//header := resp.Header
	body, err := network.GetResponseBody(resp, false)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	result := utils.ParseJson(body)
	//fmt.Println(utils.ToJson(result))
	data := result["data"]
	var songHashId = ""
	if data != nil && data.(common.MapType)["lists"] != nil && data.(common.MapType)["lists"].(common.SliceType) != nil && len(data.(common.MapType)["lists"].(common.SliceType)) > 0 {
		for _, matched := range data.(common.MapType)["lists"].(common.SliceType) {
			if matched != nil && matched.(common.MapType)["FileHash"] != nil && strings.Contains(keyword, matched.(common.MapType)["SingerName"].(string)) {
				songHashId = matched.(common.MapType)["FileHash"].(string)
				searchSong.Artist = matched.(common.MapType)["SingerName"].(string)
				searchSong.Name = matched.(common.MapType)["SongName"].(string)
				//fmt.Println(utils.ToJson(matched))
				break
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
		body, err = network.GetResponseBody(resp, false)
		songData := utils.ParseJson(body)
		data = songData["data"]
		//fmt.Println(utils.ToJson(songData))
		switch data.(type) {
		case common.MapType:
			if data.(common.MapType)["play_url"] != nil {
				songUrl := data.(common.MapType)["play_url"].(string)
				if strings.Index(songUrl, "http") == 0 {
					searchSong.Url = songUrl
					searchSong.Size, _ = data.(common.MapType)["filesize"].(json.Number).Int64()
					searchSong.Br, _ = data.(common.MapType)["bitrate"].(int)
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
	rand.Seed(time.Now().UnixNano())
	num := uint64((1 + rand.Float64()) * 0x10000)
	num = num | 0
	return strconv.FormatUint(num, 16)[1:]
}
