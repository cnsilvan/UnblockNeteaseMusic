package kuwo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider/base"
	"html"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
)

type KuWo struct{}

const (
	SearchSongURL = "http://search.kuwo.cn/r.s?&correct=1&stype=comprehensive&encoding=utf8&rformat=json&mobi=1&show_copyright_off=1&searchapi=6&all=%s"
)

var blockSongUrl = map[string]json.Number{
	"2914632520.mp3":  "7",
	"2272659253.mp3":  "7",
	"3992993176.flac": "7",
}

var lock sync.Mutex

func (m *KuWo) SearchSong(song common.SearchSong) (songs []*common.Song) {
	song = base.PreSearchSong(song)
	keyWordList := utils.Combination(song.ArtistList)
	wg := sync.WaitGroup{}
	for _, v := range keyWordList {
		wg.Add(1)
		// use goroutine to deal multiple request
		go func(word string) {
			defer wg.Done()
			keyWord := song.Name
			if len(word) != 0 {
				keyWord = fmt.Sprintf("%s %s", song.Name, word)
			}
			//token := getToken(keyWord)
			header := make(http.Header, 4)
			//header["referer"] = append(header["referer"], "http://www.kuwo.cn/search/list?key="+url.QueryEscape(keyWord))
			//header["csrf"] = append(header["csrf"], token)
			//header["cookie"] = append(header["cookie"], "kw_token="+token)
			searchUrl := fmt.Sprintf(SearchSongURL, keyWord)
			result, err := base.Fetch(searchUrl, nil, header, true)
			if err != nil {
				log.Println(err)
				return
			}
			list, ok := result["content"].(common.SliceType)[1].(common.MapType)["musicpage"].(common.MapType)["abslist"].(common.SliceType)
			if ok {
				if ok && len(list) > 0 {
					listLength := len(list)
					maxIndex := listLength/2 + 1
					if maxIndex > 5 {
						maxIndex = 5
					}
					for index, matched := range list {
						if index >= maxIndex { //kuwo list order by score default
							break
						}
						kuWoSong, ok := matched.(common.MapType)
						if ok {
							musicRid, ok := kuWoSong["MUSICRID"].(string)
							if ok {
								rids := strings.Split(musicRid, "_")
								rid := rids[len(rids)-1]
								songResult := &common.Song{}
								singerName := html.UnescapeString(kuWoSong["ARTIST"].(string))
								songName := html.UnescapeString(kuWoSong["SONGNAME"].(string))
								//musicSlice := strings.Split(musicrid, "_")
								//musicId := musicSlice[len(musicSlice)-1]
								songResult.PlatformUniqueKey = kuWoSong
								songResult.PlatformUniqueKey["UnKeyWord"] = song.Keyword
								songResult.Source = "kuwo"
								songResult.PlatformUniqueKey["header"] = header
								songResult.PlatformUniqueKey["musicId"] = rid
								songResult.Id = rid
								if len(songResult.Id) > 0 {
									songResult.Id = string(common.KuWoTag) + songResult.Id
								}
								songResult.Name = songName
								songResult.Artist = singerName
								songResult.AlbumName = html.UnescapeString(kuWoSong["ALBUM"].(string))
								songResult.Artist = strings.ReplaceAll(singerName, " ", "")
								songResult.MatchScore, ok = base.CalScore(song, songName, singerName, index, maxIndex)
								if !ok {
									continue
								}
								// protect slice thread safe
								lock.Lock()
								songs = append(songs, songResult)
								lock.Unlock()
							}
						}
					}
				}
			}
		}(v)
	}
	wg.Wait()
	return base.AfterSearchSong(song, songs)
}
func (m *KuWo) GetSongUrl(searchSong common.SearchMusic, song *common.Song) *common.Song {
	if id, ok := song.PlatformUniqueKey["musicId"]; ok {
		if musicId, ok := id.(string); ok {
			if httpHeader, ok := song.PlatformUniqueKey["header"]; ok {
				if header, ok := httpHeader.(http.Header); ok {
					header["user-agent"] = append(header["user-agent"], "okhttp/3.10.0")
					format := "flac|mp3"
					br := ""
					retry := true
					for retry {
						retry = false
						switch searchSong.Quality {
						case common.Standard:
							format = "mp3"
							br = "&br=128kmp3"
						case common.Higher:
							format = "mp3"
							br = "&br=192kmp3"
						case common.ExHigh:
							format = "mp3"
						case common.Lossless:
							format = "flac|mp3"
						default:
							format = "flac|mp3"
						}

						clientRequest := network.ClientRequest{
							Method:               http.MethodGet,
							ForbiddenEncodeQuery: true,
							RemoteUrl:            "http://mobi.kuwo.cn/mobi.s?f=kuwo&q=" + base64.StdEncoding.EncodeToString(Encrypt([]byte("corp=kuwo&p2p=1&type=convert_url2&sig=0&format="+format+"&rid="+musicId+br))),
							Header:               header,
							Proxy:                true,
						}
						resp, err := network.Request(&clientRequest)
						if err != nil {
							log.Println(err)
							return song
						}
						defer resp.Body.Close()
						body, err := network.GetResponseBody(resp, false)
						reg := regexp.MustCompile(`http[^\s$"]+`)
						address := string(body)
						params := reg.FindStringSubmatch(address)
						if len(params) > 0 {
							u, _ := url.Parse(params[0])
							if _, ok := blockSongUrl[path.Base(u.Path)]; ok {
								log.Println(song.PlatformUniqueKey["UnKeyWord"].(string) + ` ` + format + "，该歌曲酷我版权保护")
								if searchSong.Quality > 0 {
									retry = true
									log.Println(song.PlatformUniqueKey["UnKeyWord"].(string) + " ，降低音质重试")
									searchSong.Quality = searchSong.Quality - 1
								} else {
									return song
								}
							} else {
								song.Url = params[0]
								return song
							}
						}
					}
				}
			}
		}
	}
	return song
}
func (m *KuWo) ParseSong(searchSong common.SearchSong) *common.Song {
	song := &common.Song{}
	songs := m.SearchSong(searchSong)
	if len(songs) > 0 {
		song = m.GetSongUrl(common.SearchMusic{Quality: searchSong.Quality}, songs[0])
	}
	return song
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
		log.Println(err)
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
