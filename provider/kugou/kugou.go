package kuwo

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
)

const (
	APIGetSongURL = "http://trackercdn.kugou.com/i/v2/?pid=2&behavior=play&cmd=25"
)

type KuGou struct{}

func (m *KuGou) SearchSong(song common.SearchSong) (songs []*common.Song) {
	song.Keyword = strings.ToUpper(song.Keyword)
	song.Name = strings.ToUpper(song.Name)
	song.ArtistsName = strings.ToUpper(song.ArtistsName)
	cookies := getCookies()
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: "http://songsearch.kugou.com/song_search_v2?keyword=" + song.Keyword + "&page=1",
		//Host:      "songsearch.kugou.com",
		Cookies: cookies,
		Proxy:   true,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		log.Println(err)
		return songs
	}
	defer resp.Body.Close()
	//header := resp.Header
	body, err := network.StealResponseBody(resp)
	if err != nil {
		log.Println(err)
		return songs
	}
	result := utils.ParseJsonV2(body)
	//log.Println(utils.ToJson(result))
	data := result["data"]
	if data != nil {
		if dMap, ok := data.(common.MapType); ok {
			if lists, ok := dMap["lists"]; ok {
				if listSlice, ok := lists.(common.SliceType); ok {
					listLength := len(listSlice)
					if listLength > 0 {
						for index, matched := range listSlice {
							if index >= listLength/2+1 || index > 9 {
								break
							}
							if kugouSong, ok := matched.(common.MapType); ok {
								if _, ok := kugouSong["FileHash"].(string); ok {
									//log.Println(utils.ToJson(kugouSong))
									songResult := &common.Song{}
									singerName, singerNameOk := kugouSong["SingerName"].(string)
									songName, songNameOk := kugouSong["SongName"].(string)
									songResult.PlatformUniqueKey = kugouSong
									songResult.PlatformUniqueKey["UnKeyWord"] = song.Keyword
									songResult.Source = "kugou"
									songResult.Name = songName
									songResult.Artist = singerName
									songResult.Artist = strings.ReplaceAll(singerName, " ", "")
									songResult.AlbumName, _ = kugouSong["AlbumName"].(string)
									songResult.Id, ok = kugouSong["ID"].(string)
									if len(songResult.Id) > 0 {
										songResult.Id = string(common.KuGouTag) + songResult.Id
									}
									if song.OrderBy == common.MatchedScoreDesc {
										if strings.Contains(songName, "伴奏") && !strings.Contains(song.Keyword, "伴奏") {
											continue
										}
										var songNameSores float32 = 0.0
										if songNameOk {
											//songNameKeys := utils.ParseSongNameKeyWord(songName)
											//log.Println("songNameKeys:", strings.Join(songNameKeys, "、"))
											//songNameSores = utils.CalMatchScores(searchSongName, songNameKeys)
											songNameSores = utils.CalMatchScoresV2(song.Name, songName, "songName")
											//log.Printf("kugou: songName:%s,searchSongName:%s,songNameSores:%v\n", songName, searchSongName, songNameSores)
										}
										var artistsNameSores float32 = 0.0
										if singerNameOk {
											//artistKeys := utils.ParseSingerKeyWord(singerName)
											//log.Println("kugou:artistKeys:", strings.Join(artistKeys, "、"))
											//artistsNameSores = utils.CalMatchScores(searchArtistsName, artistKeys)
											artistsNameSores = utils.CalMatchScoresV2(song.ArtistsName, singerName, "singerName")
											//log.Printf("kugou: singerName:%s,searchArtistsName:%s,artistsNameSores:%v\n", singerName, searchArtistsName, artistsNameSores)
										}
										songMatchScore := songNameSores*0.6 + artistsNameSores*0.4
										songResult.MatchScore = songMatchScore
										//log.Println("kugou:songMatchScore:", songMatchScore)
									} else if song.OrderBy == common.PlatformDefault {

									}
									songs = append(songs, songResult)

								}

							}

						}
					}

				}

			}
		}
	}
	if song.OrderBy == common.MatchedScoreDesc && len(songs) > 1 {
		sort.Sort(common.SongSlice(songs))
	}
	if song.Limit > 0 && len(songs) > song.Limit {
		songs = songs[:song.Limit]
	}
	return songs
}
func (m *KuGou) GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song {
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
	//log.Println(utils.ToJson(kugouSearchSong))
	if fileHash, ok := song.PlatformUniqueKey["FileHash"].(string); ok && len(fileHash) > 0 {
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
		//log.Println(clientRequest.RemoteUrl)
		resp, err := network.Request(&clientRequest)
		if err != nil {
			log.Println(err)
			return song
		}
		defer resp.Body.Close()
		body, err := network.StealResponseBody(resp)
		songData := utils.ParseJsonV2(body)
		//log.Println(utils.ToJson(songData))
		status, ok := songData["status"].(json.Number)
		if !ok || status.String() != "1" {
			log.Println(song.PlatformUniqueKey["UnKeyWord"].(string) + "，该歌曲酷狗版权保护")
			log.Println(utils.ToJson(songData))
			return song
		}
		songUrls, ok := songData["url"].(common.SliceType)
		if ok && len(songUrls) > 0 {
			songUrl, ok := songUrls[0].(string)
			if ok && strings.Index(songUrl, "http") == 0 {
				song.Url = songUrl
				if br, ok := songData["bitRate"]; ok {
					switch br.(type) {
					case json.Number:
						song.Br, _ = strconv.Atoi(br.(json.Number).String())
					case int:
						song.Br = br.(int)
					}
				}
				return song

			}
		}
		//log.Println(utils.ToJson(data))
	}
	return song
}

func (m *KuGou) ParseSong(searchSong common.SearchSong) *common.Song {
	song := &common.Song{}
	songs := m.SearchSong(searchSong)
	if len(songs) > 0 {
		song = m.GetSongUrl(common.SearchMusic{Quality: searchSong.Quality}, songs[0])
	}
	return song
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
