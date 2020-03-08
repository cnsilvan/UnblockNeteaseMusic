package kuwo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
	"net/http"
	"strconv"
	"strings"
)

const (
	APIGetSongURL = "http://trackercdn.kugou.com/i/v2/?pid=2&behavior=play&cmd=25"
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
		Proxy:     true,
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
	var kugouSearchSong common.MapType = nil
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
							if _, ok := kugouSong["FileHash"].(string); ok {
								singerName, singerNameOk := kugouSong["SingerName"].(string)
								songName, songNameOk := kugouSong["SongName"].(string)
								if strings.Contains(songName, "伴奏") && !strings.Contains(searchSongName, "伴奏") {
									continue
								}
								var songNameSores float32 = 0.0
								if songNameOk {
									//songNameKeys := utils.ParseSongNameKeyWord(songName)
									//fmt.Println("songNameKeys:", strings.Join(songNameKeys, "、"))
									//songNameSores = utils.CalMatchScores(searchSongName, songNameKeys)
									songNameSores = utils.CalMatchScoresV2(searchSongName, songName, "songName")
									//fmt.Printf("kugou: songName:%s,searchSongName:%s,songNameSores:%v\n", songName, searchSongName, songNameSores)
								}
								var artistsNameSores float32 = 0.0
								if singerNameOk {
									//artistKeys := utils.ParseSingerKeyWord(singerName)
									//fmt.Println("kugou:artistKeys:", strings.Join(artistKeys, "、"))
									//artistsNameSores = utils.CalMatchScores(searchArtistsName, artistKeys)
									artistsNameSores = utils.CalMatchScoresV2(searchArtistsName, singerName, "singerName")
									//fmt.Printf("kugou: singerName:%s,searchArtistsName:%s,artistsNameSores:%v\n", singerName, searchArtistsName, artistsNameSores)
								}
								songMatchScore := songNameSores*0.6 + artistsNameSores*0.4
								//fmt.Println("kugou:songMatchScore:", songMatchScore)
								if songMatchScore > searchSong.MatchScore {
									searchSong.MatchScore = songMatchScore
									//songHashId = fileHash
									kugouSearchSong = kugouSong
									searchSong.Name = songName
									searchSong.Artist = singerName
									searchSong.Artist = strings.ReplaceAll(singerName, " ", "")
									//fmt.Println(utils.ToJson(searchSong))
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

	//'album_audio_id' => '32028735',
	//	'album_id' => '958812',
	//	'appid' => '1000',
	//	'area_code' => '1',
	//	'authType' => '1',
	//	'behavior' => 'play',
	//// 取现行时间戳
	//	'clienttime' => '1568759380',
	//	'clientver' => '9315',
	//	'cmd' => '26',
	//	'dfid' => '10rEiE2OosQU4Tx1VT14QDMm',
	//	'hash' => '02394f376798e3683d12ce4d1e39bae1',
	//	'key' => '66c71aac75337a20b786b92b31462ea7',
	//	'mid' => '5608f434bafb2712c78357a0c75f142a36dc5861',
	//	'module' => 'collection',
	//	'mtype' => '1',
	//	'pid' => '3',
	//	'pidversion' => '3001',
	//	'ptype' => '0',
	//// 酷狗token
	//	'token' => '打码',
	//// 酷狗uid
	//	'userid' => '打码',
	//	'version' => '9315',
	//	'vipType' => '6',
	//	'signature' => '82201ea785e8f05e4edbd9f1f6d084cf',
	//fmt.Println(utils.ToJson(kugouSearchSong))
	if fileHash, ok := kugouSearchSong["FileHash"].(string); ok && len(fileHash) > 0 {
		//audioId := "000"
		//albumId := "000"
		//if audioId1, ok := kugouSearchSong["Audioid"].(json.Number); ok {
		//	audioId = audioId1.String()
		//}
		//albumId, ok = kugouSearchSong["AlbumID"].(string)
		//http://trackercdnbj.kugou.com/i/v2/?album_audio_id=99121191&behavior=play&cmd=25
		//&album_id=6960309&hash=b5a2d566c9de70422f5e5e7203054219&userid=0&pid=2
		//&version=9108&area_code=1&appid=1005&key=407732fc325852538ca836581fe4e370&pidversion=3001&with_res_tag=1
		//mid := "0"
		//userid := "0"

		clientRequest := network.ClientRequest{
			Method:    http.MethodGet,
			RemoteUrl: APIGetSongURL + "&hash=" + fileHash + "&key=" + utils.MD5([]byte(fileHash+"kgcloudv2")),
			Host:      "trackercdnbj.kugou.com",
			//Cookies:              cookies,
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
		//fmt.Println(utils.ToJson(songData))
		status, ok := songData["status"].(json.Number)
		if !ok || status.String() != "1" {
			fmt.Println(keyword + "，该歌曲酷狗版权保护")
			fmt.Println(utils.ToJson(songData))
			return searchSong
		}
		songUrls, ok := songData["url"].(common.SliceType)
		if ok && len(songUrls) > 0 {
			songUrl, ok := songUrls[0].(string)
			if ok && strings.Index(songUrl, "http") == 0 {
				searchSong.Url = songUrl
				if br, ok := songData["bitRate"]; ok {
					switch br.(type) {
					case json.Number:
						searchSong.Br, _ = strconv.Atoi(br.(json.Number).String())
					case int:
						searchSong.Br = br.(int)
					}
				}
				return searchSong

			}
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
