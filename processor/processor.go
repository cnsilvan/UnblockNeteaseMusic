package processor

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"UnblockNeteaseMusic/host"
	"io"
	"io/ioutil"
	"net/http"
	"UnblockNeteaseMusic/network"
	"UnblockNeteaseMusic/processor/crypto"
	"UnblockNeteaseMusic/provider"
	"regexp"
	"strings"
	"UnblockNeteaseMusic/utils"
)

var (
	eApiKey     = "e82ckenh8dichen8"
	linuxApiKey = "rFgB&h#%2?^eDg:Q"
	///api/song/enhance/player/url
	///eapi/mlivestream/entrance/playlist/get
	Path = map[string]int{
		"/api/v3/playlist/detail":            1,
		"/api/v3/song/detail":                1,
		"/api/v6/playlist/detail":            1,
		"/api/album/play":                    1,
		"/api/artist/privilege":              1,
		"/api/album/privilege":               1,
		"/api/v1/artist":                     1,
		"/api/v1/artist/songs":               1,
		"/api/artist/top/song":               1,
		"/api/v1/album":                      1,
		"/api/album/v3/detail":               1,
		"/api/playlist/privilege":            1,
		"/api/song/enhance/player/url":       1,
		"/api/song/enhance/player/url/v1":    1,
		"/api/song/enhance/download/url":     1,
		"/batch":                             1,
		"/api/batch":                         1,
		"/api/v1/search/get":                 1,
		"/api/cloudsearch/pc":                1,
		"/api/v1/playlist/manipulate/tracks": 1,
		"/api/song/like":                     1,
		"/api/v1/play/record":                1,
		"/api/playlist/v4/detail":            1,
		"/api/v1/radio/get":                  1,
		"/api/v1/discovery/recommend/songs":  1,
		"/api/cloudsearch/get/web":           1,
		"/api/song/enhance/privilege":        1,
	}
)

type Netease struct {
	Path      string
	Params    map[string]interface{}
	JsonBody  map[string]interface{}
	Web       bool
	Encrypted bool
}
type MapType = map[string]interface{}
type SliceType = []interface{}

func RequestBefore(request *http.Request) *Netease {
	netease := &Netease{Path: request.URL.Path}
	if request.Method == http.MethodPost && (strings.Contains(netease.Path, "/eapi/") || strings.Contains(netease.Path, "/api/linux/forward")) {
		request.Header.Del("x-napm-retry")
		request.Header.Set("X-Real-IP", "118.66.66.66")
		requestBody, _ := ioutil.ReadAll(request.Body)
		requestHold := ioutil.NopCloser(bytes.NewBuffer(requestBody))
		request.Body = requestHold
		pad := make([]byte, 0)
		if matched, _ := regexp.Match("/%0 +$/", requestBody); matched {
			pad = requestBody
		}
		if netease.Path == "/api/linux/forward" {
			requestBodyH := make([]byte, len(requestBody))
			length, _ := hex.Decode(requestBodyH, requestBody[8:len(requestBody)-len(pad)])
			decryptECBBytes, _ := crypto.AesDecryptECB(requestBodyH[:length], []byte(linuxApiKey))
			var result MapType
			d := utils.JSON.NewDecoder(bytes.NewReader(decryptECBBytes))
			d.UseNumber()
			d.Decode(&result)
			if utils.Exist("url", result) && utils.Exist("path", result["url"].(MapType)) {
				netease.Path = result["url"].(MapType)["path"].(string)
			}
			netease.Params = utils.ParseJson(bytes.NewBufferString(result["params"].(string)).Bytes())
			fmt.Println("forward")
			//fmt.Printf("path:%s \nparams:%s\n", netease.Path, netease.Params)
		} else {
			requestBodyH := make([]byte, len(requestBody))
			length, _ := hex.Decode(requestBodyH, requestBody[7:len(requestBody)-len(pad)])
			decryptECBBytes, _ := crypto.AesDecryptECB(requestBodyH[:length], []byte(eApiKey))
			decryptString := string(decryptECBBytes)
			data := strings.Split(decryptString, "-36cd479b6b5-")
			netease.Path = data[0]
			netease.Params = utils.ParseJson(bytes.NewBufferString(data[1]).Bytes())
			//fmt.Printf("path:%s \nparams:%s\n", netease.Path, netease.Params)
		}

	} else if strings.Index(netease.Path, "/weapi/") == 0 || strings.Index(netease.Path, "/api/") == 0 {
		request.Header.Set("X-Real-IP", "118.66.66.66")
		netease.Web = true
		netease.Path = utils.ReplaceAll(netease.Path, `^\/weapi\/`, "/api/")
		netease.Path = utils.ReplaceAll(netease.Path, `\?.+$`, "")
		netease.Path = utils.ReplaceAll(netease.Path, `\/\d*$`, "")
	} else if strings.Contains(netease.Path, "package") {

	}

	return netease
}
func Request(request *http.Request, remoteUrl string) (*http.Response, error) {
	clientRequest := network.ClientRequest{
		Method:    request.Method,
		RemoteUrl: remoteUrl,
		Host:      request.Host,
		Header:    request.Header,
		Body:      request.Body,
		Proxy:     true,
	}
	return network.Request(&clientRequest)
}
func RequestAfter(request *http.Request, response *http.Response, netease *Netease) {
	run := false
	if _, ok := Path[netease.Path]; ok {
		run = true
	}
	if run && response.StatusCode == 200 {
		encode := response.Header.Get("Content-Encoding")
		enableGzip := false
		if len(encode) > 0 && (strings.Contains(encode, "gzip") || strings.Contains(encode, "deflate")) {
			enableGzip = true
		}
		body, _ := ioutil.ReadAll(response.Body)
		tmpBody := make([]byte, len(body))
		copy(tmpBody, body)
		if len(body) > 0 {
			decryptECBBytes := body
			if enableGzip {
				r, _ := gzip.NewReader(bytes.NewReader(decryptECBBytes))
				defer r.Close()
				decryptECBBytes, _ = ioutil.ReadAll(r)
			}
			//fmt.Println(string(decryptECBBytes), netease)
			decryptECBBytes, encrypted := crypto.AesDecryptECB(decryptECBBytes, []byte(eApiKey))
			netease.Encrypted = encrypted
			result := utils.ParseJson(decryptECBBytes)
			netease.JsonBody = result
			//fmt.Println(utils.ToJson(result))
			modified := false
			code := netease.JsonBody["code"].(json.Number).String()
			//fmt.Println(netease)
			if !netease.Web && (code == "401" || code == "512") && strings.Contains(netease.Path, "manipulate") {
				modified = tryCollect(netease, request)
			} else if !netease.Web && (code == "401" || code == "512") && strings.EqualFold(netease.Path, "/api/song/like") {
				modified = tryLike(netease, request)
			} else if strings.Contains(netease.Path, "url") {
				modified = tryMatch(netease)
				fmt.Println(utils.ToJson(netease.JsonBody))
			}
			if processMapJson(netease.JsonBody) || modified {
				response.Header.Del("transfer-encoding")
				response.Header.Del("content-encoding")
				response.Header.Del("content-length")
				//netease.JsonBody = netease.JsonBody
				fmt.Println("NeedRepackage")
				modifiedJson, _ := utils.JSON.Marshal(netease.JsonBody)
				//fmt.Println(netease)
				//fmt.Println(string(modifiedJson))
				if netease.Encrypted {
					modifiedJson = crypto.AesEncryptECB(modifiedJson, []byte(eApiKey))
				}
				response.Body = ioutil.NopCloser(bytes.NewBuffer(modifiedJson))
			} else {
				fmt.Println("NotNeedRepackage")
				responseHold := ioutil.NopCloser(bytes.NewBuffer(tmpBody))
				response.Body = responseHold
			}

		} else {
			responseHold := ioutil.NopCloser(bytes.NewBuffer(tmpBody))
			response.Body = responseHold
		}
	} else {
		//fmt.Println("Not Process")
	}
}
func tryCollect(netease *Netease, request *http.Request) bool {
	modified := false
	//fmt.Println(utils.ToJson(netease))
	if utils.Exist("trackIds", netease.Params) {
		trackId := ""
		switch netease.Params["trackIds"].(type) {
		case string:
			var result SliceType
			d := utils.JSON.NewDecoder(bytes.NewReader(bytes.NewBufferString(netease.Params["trackIds"].(string)).Bytes()))
			d.UseNumber()
			d.Decode(&result)
			trackId = result[0].(string)
		case SliceType:
			trackId = netease.Params["trackIds"].(SliceType)[0].(json.Number).String()
		}
		pid := netease.Params["pid"].(string)
		op := netease.Params["op"].(string)
		proxyRemoteHost := host.ProxyDomain["music.163.com"]
		clientRequest := network.ClientRequest{
			Method:    http.MethodPost,
			Host:      "music.163.com",
			RemoteUrl: "http://" + proxyRemoteHost + "/api/playlist/manipulate/tracks",
			Header:    request.Header,
			Body:      ioutil.NopCloser(bytes.NewBufferString("trackIds=[" + trackId + "," + trackId + "]&pid=" + pid + "&op=" + op)),
		}
		resp, err := network.Request(&clientRequest)
		if err != nil {
			return modified
		}
		body, err := network.GetResponseBody(resp, false)
		if err != nil {
			return modified
		}
		netease.JsonBody = utils.ParseJson(body)
		modified = true
	}
	return modified
}
func tryLike(netease *Netease, request *http.Request) bool {
	//fmt.Println("try like")
	modified := false
	if utils.Exist("trackId", netease.Params) {
		trackId := netease.Params["trackId"].(string)
		proxyRemoteHost := host.ProxyDomain["music.163.com"]
		clientRequest := network.ClientRequest{
			Method:    http.MethodGet,
			Host:      "music.163.com",
			RemoteUrl: "http://" + proxyRemoteHost + "/api/v1/user/info",
			Header:    request.Header}
		resp, err := network.Request(&clientRequest)
		if err != nil {
			return modified
		}
		body, err := network.GetResponseBody(resp, false)
		if err != nil {
			return modified
		}
		jsonBody := utils.ParseJson(body)
		if utils.Exist("userPoint", jsonBody) && utils.Exist("userId", jsonBody["userPoint"].(MapType)) {
			userId := jsonBody["userPoint"].(MapType)["userId"].(json.Number).String()
			clientRequest.RemoteUrl = "http://" + proxyRemoteHost + "/api/user/playlist?uid=" + userId + "&limit=1"
			resp, err = network.Request(&clientRequest)
			if err != nil {
				return modified
			}
			body, err = network.GetResponseBody(resp, false)
			if err != nil {
				return modified
			}
			jsonBody = utils.ParseJson(body)
			if utils.Exist("playlist", jsonBody) {
				pid := jsonBody["playlist"].(SliceType)[0].(MapType)["id"].(json.Number).String()
				clientRequest.Method = http.MethodPost
				clientRequest.RemoteUrl = "http://" + proxyRemoteHost + "/api/playlist/manipulate/tracks"
				clientRequest.Body = ioutil.NopCloser(bytes.NewBufferString("trackIds=[" + trackId + "," + trackId + "]&pid=" + pid + "&op=add"))
				resp, err = network.Request(&clientRequest)
				if err != nil {
					return modified
				}
				body, err = network.GetResponseBody(resp, false)
				if err != nil {
					return modified
				}
				jsonBody = utils.ParseJson(body)
				code := jsonBody["code"].(json.Number).String()
				if code == "200" || code == "502" {
					netease.JsonBody = make(MapType)
					netease.JsonBody["code"] = 200
					netease.JsonBody["playlistId"] = pid
					modified = true
				}
			}
		}
	}

	return modified
}
func tryMatch(netease *Netease) bool {
	fmt.Println(netease.Path)
	modified := false
	jsonBody := netease.JsonBody
	if value, ok := jsonBody["data"]; ok {
		switch value.(type) {
		case SliceType:
			if strings.Contains(netease.Path, "download") {
				for index, data := range value.(SliceType) {
					if index == 0 {
						modified = searchGreySong(data.(MapType), netease) || modified
						break
					}
				}
			} else {
				modified = searchGreySongs(value.(SliceType), netease) || modified
			}
		case MapType:
			modified = searchGreySong(value.(MapType), netease) || modified
		default:
		}
	}
	//modifiedJson, _ := json.Marshal(jsonBody)
	//fmt.Println(string(modifiedJson))
	return modified
}
func searchGreySongs(data SliceType, netease *Netease) bool {
	modified := false
	for _, value := range data {
		switch value.(type) {
		case MapType:
			modified = searchGreySong(value.(MapType), netease) || modified
		}
	}
	return modified
}
func searchGreySong(data MapType, netease *Netease) bool {
	modified := false
	if data["url"] == nil {
		data["flag"] = 0
		songId := data["id"].(json.Number).String()
		song := provider.Find(songId)
		haveSongMd5 := false
		if song.Size > 0 {
			modified = true
			if song.Br == 999000 {
				data["type"] = "flac"
			} else {
				data["type"] = "mp3"
			}
			data["encodeType"] = data["type"] //web
			data["level"] = "standard"        //web
			data["fee"] = 8                   //web
			data["url"] = song.Url
			if len(song.Md5) > 0 {
				data["md5"] = song.Md5
				haveSongMd5 = true
			} else {
				h := md5.New()
				h.Write([]byte(song.Url))
				data["md5"] = hex.EncodeToString(h.Sum(nil))
				haveSongMd5 = false
			}
			if song.Br > 0 {
				data["br"] = song.Br
			} else {
				data["br"] = 128000
			}
			data["size"] = song.Size
			data["freeTrialInfo"] = nil
			data["code"] = 200
			if strings.Contains(netease.Path, "download") { //calculate the file md5
				if !haveSongMd5 {
					data["md5"] = calculateSongMd5(songId, song.Url)
				}
			} else if !haveSongMd5 {
				go calculateSongMd5(songId, song.Url)
			}
		}

	}
	return modified
}
func calculateSongMd5(songId string, songUrl string) string {
	songMd5 := ""
	clientRequest := network.ClientRequest{
		Method:    http.MethodGet,
		RemoteUrl: songUrl,
	}
	resp, err := network.Request(&clientRequest)
	if err != nil {
		fmt.Println(err)
		return songMd5
	}
	r := bufio.NewReader(resp.Body)
	h := md5.New()
	_, err = io.Copy(h, r)
	if err != nil {
		fmt.Println(err)
		return songMd5
	}
	songMd5 = hex.EncodeToString(h.Sum(nil))
	provider.UpdateCacheMd5(songId, songMd5)
	//fmt.Println("calculateSongMd5 songId:", songId, ",songUrl:", songUrl, ",md5:", songMd5)
	return songMd5
}
func processSliceJson(jsonSlice SliceType) bool {
	needModify := false
	for _, value := range jsonSlice {
		switch value.(type) {
		case MapType:
			needModify = processMapJson(value.(MapType)) || needModify

		case SliceType:
			needModify = processSliceJson(value.(SliceType)) || needModify

		default:
			//fmt.Printf("index(%T):%v\n", index, index)
			//fmt.Printf("value(%T):%v\n", value, value)
		}
	}
	return needModify
}
func processMapJson(jsonMap MapType) bool {
	needModify := false
	if utils.Exists([]string{"st", "subp", "pl", "dl"}, jsonMap) {
		if v, _ := jsonMap["st"]; v.(json.Number).String() != "0" {
			//open grep song
			jsonMap["st"] = 0
			needModify = true
		}
		if v, _ := jsonMap["subp"]; v.(json.Number).String() != "1" {
			jsonMap["subp"] = 1
			needModify = true
		}
		if v, _ := jsonMap["pl"]; v.(json.Number).String() == "0" {
			jsonMap["pl"] = 320000
			needModify = true
		}
		if v, _ := jsonMap["dl"]; v.(json.Number).String() == "0" {
			jsonMap["dl"] = 320000
			needModify = true
		}
	}
	for _, value := range jsonMap {
		switch value.(type) {
		case MapType:
			needModify = processMapJson(value.(MapType)) || needModify
		case SliceType:
			needModify = processSliceJson(value.(SliceType)) || needModify
		default:
			//if key == "fee" && value.(json.Number).String() != "0" {
			//	jsonMap[key] = 0
			//	needModify = true
			//}
		}
	}
	return needModify
}
