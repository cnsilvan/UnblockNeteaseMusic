package processor

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"processor/crypto"
	"provider"
	"regexp"
	"strings"
	"utils"
)

var (
	eApiKey     = "e82ckenh8dichen8"
	linuxApiKey = "rFgB&h#%2?^eDg:Q"
)

type Netease struct {
	Path     string
	Params   string
	JsonBody map[string]interface{}
}

type MapType = map[string]interface{}
type SliceType = []interface{}

func DecodeRequestBody(request *http.Request) *Netease {
	netease := &Netease{Path: request.RequestURI}
	requestBody, _ := ioutil.ReadAll(request.Body)
	//fmt.Println(string(requestBody))
	requestHold := ioutil.NopCloser(bytes.NewBuffer(requestBody))
	request.Body = requestHold
	pad := make([]byte, 0)
	if matched, _ := regexp.Match("/%0 +$/", requestBody); matched {
		pad = requestBody
	}
	if netease.Path == "/api/linux/forward" {
		requestBodyH := make([]byte, len(requestBody))
		length, _ := hex.Decode(requestBodyH, requestBody[8:len(requestBody)-len(pad)])
		decryptECBBytes := crypto.AesDecryptECB(requestBodyH[:length], []byte(linuxApiKey))
		var result MapType
		d := json.NewDecoder(bytes.NewReader(decryptECBBytes))
		d.UseNumber()
		d.Decode(&result)
		if utils.Exist("url", result) && utils.Exist("path", result["url"].(MapType)) {
			netease.Path = result["url"].(MapType)["path"].(string)
		}
		netease.Params = result["params"].(string)
		fmt.Println("forward")
		//fmt.Printf("path:%s \nparams:%s\n", netease.Path, netease.Params)
	} else {
		requestBodyH := make([]byte, len(requestBody))
		length, _ := hex.Decode(requestBodyH, requestBody[7:len(requestBody)-len(pad)])
		decryptECBBytes := crypto.AesDecryptECB(requestBodyH[:length], []byte(eApiKey))
		decryptString := string(decryptECBBytes)
		data := strings.Split(decryptString, "-36cd479b6b5-")
		netease.Path = data[0]
		netease.Params = data[1]
		//fmt.Printf("path:%s \nparams:%s\n", netease.Path, netease.Params)
	}

	return netease
}
func DecodeResponseBody(response *http.Response, netease *Netease) {
	if response.StatusCode == 200 {
		encode := response.Header.Get("Content-Encoding")
		enableGzip := false
		if len(encode) > 0 && (strings.Contains(encode, "gzip") || strings.Contains(encode, "deflate")) {
			enableGzip = true
		}
		body, _ := ioutil.ReadAll(response.Body)
		if len(body) > 0 {
			decryptECBBytes := body
			if enableGzip {
				r, _ := gzip.NewReader(bytes.NewReader(decryptECBBytes))
				defer r.Close()
				decryptECBBytes, _ = ioutil.ReadAll(r)
			}
			decryptECBBytes = crypto.AesDecryptECB(decryptECBBytes, []byte(eApiKey))
			result := utils.ParseJson(decryptECBBytes)
			modified := false
			if strings.Contains(netease.Path, "manipulate") {

			} else if strings.EqualFold(netease.Path, "/api/song/like") {
			} else if strings.Contains(netease.Path, "url") {
				fmt.Println(netease.Path)
				if value, ok := result["data"]; ok {
					switch value.(type) {
					case SliceType:
						if strings.Contains(netease.Path, "download") {
							for index, data := range value.(SliceType) {
								if index == 0 {
									modified = searchGreySong(data.(MapType)) || modified
									break
								}
							}
						} else {
							modified = searchGreySongs(value.(SliceType)) || modified
						}
					case MapType:
						modified = searchGreySong(value.(MapType)) || modified
					default:
					}
				}
				modifiedJson, _ := json.Marshal(result)
				fmt.Println(string(modifiedJson))
			}
			if processMapJson(result) || modified {
				response.Header.Del("transfer-encoding")
				response.Header.Del("content-encoding")
				response.Header.Del("content-length")
				netease.JsonBody = result
				fmt.Println("NeedRepackage")
				modifiedJson, _ := json.Marshal(result)
				//fmt.Println(string(modifiedJson))
				encryptedJson := crypto.AesEncryptECB(modifiedJson, []byte(eApiKey))
				response.Body = ioutil.NopCloser(bytes.NewBuffer(encryptedJson))
			} else {
				//fmt.Println(string(body))
				responseHold := ioutil.NopCloser(bytes.NewBuffer(body))
				response.Body = responseHold
			}

		} else {
			responseHold := ioutil.NopCloser(bytes.NewBuffer(body))
			response.Body = responseHold
		}
	} else {
		//body, _ := ioutil.ReadAll(response.Body)
		//responseHold := ioutil.NopCloser(bytes.NewBuffer(body))
		//response.Body = responseHold
		//fmt.Println(string(body))
	}
}
func searchGreySongs(data SliceType) bool {
	modified := false
	for _, value := range data {
		switch value.(type) {
		case MapType:
			modified = searchGreySong(value.(MapType)) || modified
		}
	}
	return modified
}
func searchGreySong(data MapType) bool {
	modified := false
	if data["url"] == nil {
		data["flag"] = 0
		song := provider.Find(data["id"].(json.Number).String())
		//modifiedJson, _ := json.Marshal(song)
		//fmt.Println(string(modifiedJson))
		if song.Size > 0 {
			modified = true
			if song.Br == 999000 {
				data["type"] = "flac"
			} else {
				data["type"] = "mp3"
			}
			data["url"] = song.Url
			if len(song.Md5) > 0 {
				data["md5"] = song.Md5
			} else {
				h := md5.New()
				h.Write([]byte(song.Url))
				data["md5"] = hex.EncodeToString(h.Sum(nil))
			}
			if song.Br > 0 {
				data["br"] = song.Br
			} else {
				data["br"] = 128000
			}
			data["size"] = song.Size
			data["freeTrialInfo"] = nil
			data["code"] = 200
		}

	}
	return modified
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
