package kuwo

import (
	"fmt"
	"net/url"
	"network"
	"strings"
	"utils"
)

func SearchSong(key map[string]interface{}) string {
	keyword := key["keyword"].(string)
	token := getToken(keyword)
	header := make(map[string]string, 3)
	header["referer"] = "http://www.kuwo.cn/search/list?key=" + url.QueryEscape(keyword)
	header["csrf"] = token
	header["cookie"] = "kw_token=" + token
	resp, err := network.Get("http://www.kuwo.cn/api/www/search/searchMusicBykeyWord?key="+keyword+"&pn=1&rn=30", "kuwo.cn", header, false)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	body, err := network.GetResponseBody(resp, false)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	result := utils.ParseJson(body)
	var musicId = ""
	if result["data"] != nil && result["data"].(map[string]interface{}) != nil && len(result["data"].(map[string]interface{})["list"].([]interface{})) > 0 {
		matched := result["data"].(map[string]interface{})["list"].([]interface{})[0]
		if matched != nil && matched.(map[string]interface{})["musicrid"] != nil {
			musicrid := matched.(map[string]interface{})["musicrid"].(string)
			musicSlice := strings.Split(musicrid, "_")
			musicId = musicSlice[len(musicSlice)-1]
		}
	}
	if len(musicId) > 0 {
		resp, err = network.Get("http://antiserver.kuwo.cn/anti.s?type=convert_url&format=mp3&response=url&rid=MUSIC_"+musicId, "antiserver.kuwo.cn", nil, false)
		if err != nil {
			fmt.Println(err)
			return ""
		}
		body, err = network.GetResponseBody(resp, false)
		address := string(body)
		if strings.Index(address, "http") == 0 {
			return address
		}

	}
	return ""

}
func getToken(keyword string) string {
	var token = ""
	resp, err := network.Get("http://kuwo.cn/search/list?key="+keyword, "kuwo.cn", nil, false)
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
