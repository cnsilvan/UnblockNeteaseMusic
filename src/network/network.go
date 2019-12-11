package network

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func Get(urlAddress string, host string, header map[string]string, tlsVerifyServerName bool) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, urlAddress, nil)
	var resp *http.Response
	if err != nil {
		fmt.Printf("NewRequest fail:%v\n", err)
		return resp, nil
	}
	req.URL.RawQuery = req.URL.Query().Encode()
	c := http.Client{}
	if len(host) > 0 {
		req.Host = host
		req.Header.Set("host", host)
		if tlsVerifyServerName {
			tr := http.Transport{
				TLSClientConfig: &tls.Config{ServerName: req.Host},
			}
			c = http.Client{
				Transport: &tr,
			}
		}
	}
	if header != nil {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-encoding", "gzip, deflate")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")
	resp, err = c.Do(req)
	if err != nil {
		fmt.Printf("http.Client.Do fail:%v\n", err)
		return resp, err
	}
	return resp, err
}
func GetResponseBody(resp *http.Response, keepBody bool) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read body fail")
		return body, err
	}
	resp.Body.Close()
	if keepBody {
		bodyHold := ioutil.NopCloser(bytes.NewBuffer(body))
		resp.Body = bodyHold
	}
	encode := resp.Header.Get("Content-Encoding")
	enableGzip := false
	if len(encode) > 0 && (strings.Contains(encode, "gzip") || strings.Contains(encode, "deflate")) {
		enableGzip = true
	}
	if enableGzip {
		resp.Header.Del("Content-Encoding")
		r, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			fmt.Println("read gzip body fail")
			return body, err
		}
		defer r.Close()
		body, err = ioutil.ReadAll(r)
		if err != nil {
			fmt.Println("read  body fail")
			return body, err
		}
	}
	return body, err
}
