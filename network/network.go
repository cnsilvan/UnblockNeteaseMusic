package network

import (
	"UnblockNeteaseMusic/common"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type Netease struct {
	Path     string
	Params   string
	JsonBody map[string]interface{}
}
type ClientRequest struct {
	Method               string
	RemoteUrl            string
	Host                 string
	ForbiddenEncodeQuery bool
	Header               http.Header
	Body                 io.Reader
	Cookies              []*http.Cookie
	Proxy                bool
}

func Request(clientRequest *ClientRequest) (*http.Response, error) {
	//fmt.Println(clientRequest.RemoteUrl)
	method := clientRequest.Method
	remoteUrl := clientRequest.RemoteUrl
	host := clientRequest.Host
	header := clientRequest.Header
	body := clientRequest.Body
	proxy := clientRequest.Proxy
	cookies := clientRequest.Cookies
	var resp *http.Response
	request, err := http.NewRequest(method, remoteUrl, body)
	if err != nil {
		fmt.Printf("NewRequest fail:%v\n", err)
		return resp, nil
	}
	if !clientRequest.ForbiddenEncodeQuery {
		request.URL.RawQuery = request.URL.Query().Encode()
	}
	if header != nil {
		request.Header = header
	}
	for _, value := range cookies {
		request.AddCookie(value)
	}
	c := http.Client{}
	tr := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if len(host) > 0 {
		request.Host = host
		request.Header.Set("host", host)
	}
	if len(request.URL.Scheme) == 0 {
		if request.TLS != nil {
			request.URL.Scheme = "https"
		} else {
			request.URL.Scheme = "http"
		}
	}
	if proxy && (request.URL.Scheme == "https" || request.TLS != nil) {
		tr.TLSClientConfig = &tls.Config{}
		// verify music.163.com certificate
		tr.TLSClientConfig.ServerName = request.Host //it doesn't contain any IP SANs
		// redirect to music.163.com will need verify self
		if _, ok := common.HostDomain[request.Host]; ok {
			tr.TLSClientConfig.InsecureSkipVerify = true
		}

	}
	c.Transport = &tr
	if !proxy {
		request.Header.Set("accept", "application/json, text/plain, */*")
		request.Header.Set("accept-encoding", "gzip, deflate")
		request.Header.Set("accept-language", "zh-CN,zh;q=0.9")
		request.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")

	}
	resp, err = c.Do(request)
	//fmt.Println(request.URL.String())
	//fmt.Println(request.Cookies())
	if err != nil {
		fmt.Println(request.Method, request.URL.String(), host)
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
