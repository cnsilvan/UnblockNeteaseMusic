package kuwo

import (
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
	"net/http"
	"net/url"
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
	defer resp.Body.Close()
	body, err := network.StealResponseBody(resp)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	result := utils.ParseJsonV2(body)
	//fmt.Println(utils.ToJson(result))
	var musicId = ""
	data, ok := result["data"].(common.MapType)
	if ok {
		list, ok := data["list"].([]interface{})
		if ok && len(list) > 0 {
			listLength := len(list)
			for index, matched := range list {
				kowoSong, ok := matched.(common.MapType)
				if ok {
					musicrid, ok := kowoSong["musicrid"].(string)
					if ok {
						singerName, singerNameOk := kowoSong["artist"].(string)
						songName, songNameOk := kowoSong["name"].(string)
						if strings.Contains(songName, "伴奏") && !strings.Contains(searchSongName, "伴奏") {
							continue
						}
						var songNameSores float32 = 0.0
						if songNameOk {
							//songNameKeys := utils.ParseSongNameKeyWord(songName)
							////fmt.Println("songNameKeys:", strings.Join(songNameKeys, "、"))
							//songNameSores = utils.CalMatchScores(searchSongName, songNameKeys)
							//fmt.Println("songNameSores:", songNameSores)
							songNameSores=utils.CalMatchScoresV2(searchSongName,songName,"songName")
						}
						var artistsNameSores float32 = 0.0
						if singerNameOk {
							singerName = strings.ReplaceAll(singerName, "&", "、")
							//artistKeys := utils.ParseSingerKeyWord(singerName)
							////fmt.Println("kuwo:artistKeys:", strings.Join(artistKeys, "、"))
							//artistsNameSores = utils.CalMatchScores(searchArtistsName, artistKeys)
							artistsNameSores=utils.CalMatchScoresV2(searchArtistsName,singerName,"singerName")
							//fmt.Println("kuwo:artistsNameSores:", artistsNameSores)
						}
						songMatchScore := songNameSores*0.6 + artistsNameSores*0.4
						//fmt.Println("kuwo:songMatchScore:", songMatchScore)
						if songMatchScore > searchSong.MatchScore {
							searchSong.MatchScore = songMatchScore
							musicSlice := strings.Split(musicrid, "_")
							musicId = musicSlice[len(musicSlice)-1]
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
		defer resp.Body.Close()
		body, err := network.GetResponseBody(resp, false)
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
