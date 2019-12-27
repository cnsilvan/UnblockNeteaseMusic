package migu

import (
	"UnblockNeteaseMusic/common"
	"UnblockNeteaseMusic/network"
	"UnblockNeteaseMusic/processor/crypto"
	"UnblockNeteaseMusic/utils"
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
)

var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC8asrfSaoOb4je+DSmKdriQJKW
VJ2oDZrs3wi5W67m3LwTB9QVR+cE3XWU21Nx+YBxS0yun8wDcjgQvYt625ZCcgin
2ro/eOkNyUOTBIbuj9CvMnhUYiR61lC1f1IGbrSYYimqBVSjpifVufxtx/I3exRe
ZosTByYp4Xwpb1+WAQIDAQAB
-----END PUBLIC KEY-----
`)
var rsaPublicKey *rsa.PublicKey

func getRsaPublicKey() (*rsa.PublicKey, error) {
	var err error = nil
	if rsaPublicKey.Size() == 0 {
		rsaPublicKey, err = crypto.ParsePublicKey(publicKey)
	}
	return rsaPublicKey, err
}
func SearchSong(key common.MapType) common.Song {
	searchSong := common.Song{
	}
	keyword := key["keyword"].(string)
	searchSongName := key["name"].(string)
	searchSongName = strings.ToUpper(searchSongName)
	searchArtistsName := key["artistsName"].(string)
	searchArtistsName = strings.ToUpper(searchArtistsName)
	header := make(http.Header, 2)
	header["origin"] = append(header["origin"], "http://music.migu.cn/")
	header["referer"] = append(header["referer"], "http://music.migu.cn/")
	clientRequest := network.ClientRequest{
		Method: http.MethodGet,
		RemoteUrl: "http://pd.musicapp.migu.cn/MIGUM2.0/v1.0/content/search_all.do?text=" + keyword +
			"&pageNo=1&pageSize=20&searchSwitch=" +
			"{\"song\":1,\"album\":0,\"singer\":0,\"tagSong\":0,\"mvSong\":0,\"songlist\":0,\"bestShow\":0}",
		Host:  "pd.musicapp.migu.cn",
		Proxy: false,
	}
	//fmt.Println(clientRequest.RemoteUrl)
	resp, err := network.Request(&clientRequest)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		return searchSong
	}
	body, err := network.StealResponseBody(resp)
	if err != nil {
		fmt.Println(err)
		return searchSong
	}
	result := utils.ParseJsonV2(body)
	//fmt.Println(utils.ToJson(result))
	var copyrightId = ""
	data, ok := result["songResultData"].(common.MapType)
	if ok {
		list, ok := data["result"].([]interface{})
		if ok && len(list) > 0 {
			listLength := len(list)
			for index, matched := range list {
				miguSong, ok := matched.(common.MapType)
				if ok {
					cId, ok := miguSong["copyrightId"].(string)
					if ok {
						singerName, singerNameOk := miguSong["singers"].([]interface{})
						songName, songNameOk := miguSong["name"].(string)
						singerNames := ""
						var songNameSores float32 = 0.0
						if songNameOk {
							songNameKeys := utils.ParseSongNameKeyWord(songName)
							//fmt.Println("songNameKeys:", strings.Join(songNameKeys, "、"))
							songNameSores = utils.CalMatchScores(searchSongName, songNameKeys)
							//fmt.Println("songNameSores:", songNameSores)
						}
						var artistsNameSores float32 = 0.0
						if singerNameOk {
							var artNames []string
							for _, art := range singerName {
								artN, ok := art.(common.MapType)["name"].(string)
								if ok {
									artNames = append(artNames, strings.TrimSpace(artN))
								}
							}
							singerNames = strings.Join(artNames, "、")
							artistKeys := utils.ParseSingerKeyWord(singerNames)
							//fmt.Println("migu:artistKeys:", strings.Join(artistKeys, "、"))
							artistsNameSores = utils.CalMatchScores(searchArtistsName, artistKeys)
							//fmt.Println("migu:artistsNameSores:", artistsNameSores)
						}
						songMatchScore := songNameSores*0.6 + artistsNameSores*0.4
						//fmt.Println("migu:songMatchScore:", songMatchScore)
						if songMatchScore > searchSong.MatchScore {
							searchSong.MatchScore = songMatchScore
							copyrightId = cId
							searchSong.Name = songName
							searchSong.Artist = singerNames
							searchSong.Artist = strings.ReplaceAll(singerNames, " ", "")
						}
					}

				}
				if index >= listLength/2 || index > 9 {
					break
				}
			}

		}
	}

	if len(copyrightId) > 0 {
		clientRequest := network.ClientRequest{
			Method:               http.MethodGet,
			RemoteUrl:            "http://music.migu.cn/v3/api/music/audioPlayer/getPlayInfo?dataType=2&" + encrypt("{\"copyrightId\":\""+copyrightId+"\"}"),
			Host:                 "music.migu.cn",
			Header:               header,
			Proxy:                false,
			ForbiddenEncodeQuery: true, //dataType first must
		}
		//fmt.Println(clientRequest.RemoteUrl)
		resp, err := network.Request(&clientRequest)
		if err != nil {
			fmt.Println(err)
			return searchSong
		}
		defer resp.Body.Close()
		body, err = network.StealResponseBody(resp)
		data := utils.ParseJsonV2(body)
		//fmt.Println(data)
		data, ok := data["data"].(common.MapType)
		if ok {
			//playInfo, ok := data["sqPlayInfo"].(common.MapType)
			//if !ok {
			playInfo, ok := data["hqPlayInfo"].(common.MapType)
			if !ok {
				playInfo, ok = data["bqPlayInfo"].(common.MapType)
			}
			//}
			if ok {
				playUrl, ok := playInfo["playUrl"].(string)
				if ok && strings.Index(playUrl, "http") == 0 {
					searchSong.Url = playUrl
					return searchSong
				}
			}
		}
	}
	return searchSong

}

func encrypt(text string) string {
	encryptedData := ""
	//fmt.Println(text)
	text = utils.ToJson(utils.ParseJson(bytes.NewBufferString(text).Bytes()))
	randomBytes, err := utils.GenRandomBytes(32)
	if err != nil {
		fmt.Println(err)
		return encryptedData
	}
	pwd := bytes.NewBufferString(hex.EncodeToString(randomBytes)).Bytes()
	salt, err := utils.GenRandomBytes(8)
	if err != nil {
		fmt.Println(err)
		return encryptedData
	}
	//key = []byte{0xaf, 0xb3, 0xac, 0x50, 0xcd, 0x1d, 0x23, 0x81, 0x58, 0x5f, 0xa7, 0xbc, 0xbd, 0x8c, 0xbe, 0x02, 0x56, 0x0f, 0xad, 0xe7, 0xd1, 0x7e, 0x2e, 0xb1, 0x14, 0x81, 0x6f, 0x27, 0xab, 0x7b, 0x6a, 0x75}
	//iv = []byte{0xfb, 0x10, 0x89, 0xb0, 0x13, 0x32, 0xf2, 0xa7, 0x02, 0x51, 0x49, 0xff, 0xbc, 0x16, 0xf0, 0x40}
	//pwd = bytes.NewBufferString("d8e28215ed6573e0fd5eb8b8ae8062542589e96f669bee6503af003c63cdfbd4").Bytes()
	//salt = []byte{0xde, 0xfc, 0x9f, 0x26, 0x29, 0xdd, 0xec, 0x37}
	key, iv := derive(pwd, salt, 256, 16)
	var data []byte
	data = append(data, bytes.NewBufferString("Salted__").Bytes()...)
	data = append(data, salt...)
	encryptedD := crypto.AesEncryptCBCWithIv(bytes.NewBufferString(text).Bytes(), key, iv)
	data = append(data, encryptedD...)
	dat := base64.StdEncoding.EncodeToString(data)
	var rsaB []byte
	pubKey, err := getRsaPublicKey()
	if err == nil {
		rsaB = crypto.RSAEncryptV2(pwd, pubKey)
	} else {
		rsaB = crypto.RSAEncrypt(pwd, publicKey)
	}
	sec := base64.StdEncoding.EncodeToString(rsaB)
	//fmt.Println("data:", dat)
	//fmt.Println("sec:", sec)
	encryptedData = "data=" + url.QueryEscape(dat)
	encryptedData = encryptedData + "&secKey=" + url.QueryEscape(sec)
	return encryptedData
}
func derive(password []byte, salt []byte, keyLength int, ivSize int) ([]byte, []byte) {
	keySize := keyLength / 8
	repeat := math.Ceil(float64(keySize+ivSize*8) / 32)
	var data []byte
	var lastData []byte
	for i := 0.0; i < repeat; i++ {
		var md5Data []byte
		md5Data = append(md5Data, lastData...)
		md5Data = append(md5Data, password...)
		md5Data = append(md5Data, salt...)
		h := md5.New()
		h.Write(md5Data)
		md5Data = h.Sum(nil)
		data = append(data, md5Data...)
		lastData = md5Data
	}
	return data[:keySize], data[keySize : keySize+ivSize]
}
