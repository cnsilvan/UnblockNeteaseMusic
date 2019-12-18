package kuwo

import (
	"UnblockNeteaseMusic/common"
	"UnblockNeteaseMusic/network"
	"UnblockNeteaseMusic/utils"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func SearchSong(key common.MapType) common.SearchSong {
	searchSong := common.SearchSong{

	}
	keyword := key["keyword"].(string)
	token := getToken(keyword)
	header := make(http.Header, 3)
	header["referer"] = append(header["referer"], "http://www.kuwo.cn/search/list?key="+url.QueryEscape(keyword))
	header["csrf"] = append(header["csrf"], token)
	header["cookie"] = append(header["cookie"], "kw_token="+token)
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: "http://www.kuwo.cn/api/www/search/searchMusicBykeyWord?key=" + keyword + "&pn=1&rn=30",
		Host:      "kuwo.cn",
		Header:    header,
		Proxy:     false,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	body, err := network.GetResponseBody(resp, false)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	result := utils.ParseJson(body)
	//fmt.Println(utils.ToJson(result))
	var musicId = ""
	if result["data"] != nil && result["data"].(common.MapType) != nil && len(result["data"].(common.MapType)["list"].([]interface{})) > 0 {
		for _, matched := range result["data"].(common.MapType)["list"].([]interface{}) {
			if matched != nil && matched.(common.MapType)["musicrid"] != nil && strings.Contains(keyword, matched.(common.MapType)["artist"].(string)) {
				musicrid := matched.(common.MapType)["musicrid"].(string)
				musicSlice := strings.Split(musicrid, "_")
				musicId = musicSlice[len(musicSlice)-1]
				searchSong.Artist = matched.(common.MapType)["artist"].(string)
				searchSong.Name = matched.(common.MapType)["name"].(string)
				fmt.Println(utils.ToJson(matched))
				break
				//songName:=matched.(common.MapType)["name"].(string)
			}
		}

	}
	if len(musicId) > 0 {
		clientRequest := network.ClientRequest{
			Method:    http.MethodGet,
			RemoteUrl: "http://antiserver.kuwo.cn/anti.s?type=convert_url&format=mp3&response=url&rid=MUSIC_" + musicId,
			Host:      "antiserver.kuwo.cn",
			Header:    header,
			Proxy:     false,
		}
		resp, err := network.Request(&clientRequest)
		if err != nil {
			fmt.Println(err)
			return searchSong
		}
		body, err = network.GetResponseBody(resp, false)
		address := string(body)
		if strings.Index(address, "http") == 0 {
			searchSong.Url = address
			return searchSong
		}

	}
	return searchSong

}
func getToken(keyword string) string {
	var token = ""
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: "http://kuwo.cn/search/list?key=" + keyword,
		Host:      "kuwo.cn",
		Header:    nil,
		Proxy:     false,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		fmt.Println(err)
		return token
	}
	defer resp.Body.Close()
	cookies := resp.Header.Get("set-cookie")
	if strings.Contains(cookies, "kw_token") {
		cookies = utils.ReplaceAll(cookies, ";.*", "")
		splitSlice := strings.Split(cookies, "=")
		token = splitSlice[len(splitSlice)-1]
	}
	return token
}
