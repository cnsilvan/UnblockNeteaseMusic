package proxy

import (
	"crypto/tls"
	"fmt"
	"host"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"processor"
	"strings"
	"time"
	"utils"
)

type HttpHandler struct{}

var (
	///api/song/enhance/player/url
	///eapi/mlivestream/entrance/playlist/get
	Path = []string{"/eapi/batch",
		"/eapi/album/privilege",
		"/eapi/cloudsearch/pc",
		"/eapi/playlist/privilege",
		"/eapi/song/enhance/player/url",
		"/eapi/v1/playlist/manipulate/tracks", //download music

	}
)

func InitProxy() {
	fmt.Println("-------------------Init Proxy-----------------------")
	//tlsBytes("server.crt", "server.key")
	go startTlsServer(":443", "./server.crt", "./server.key", &HttpHandler{})
	startServer(":80", &HttpHandler{})
}
func (h *HttpHandler) ServeHTTP(resp http.ResponseWriter, request *http.Request) {
	fmt.Println(request.Host)
	fmt.Println(request.URL.String())
	fmt.Println(request.RequestURI)
	if proxyDomain, ok := host.ProxyDomain[request.Host]; ok {
		scheme := "http://"
		if request.TLS != nil {
			scheme = "https://"
		}
		requestURI := request.RequestURI
		request.Header.Del("x-napm-retry")
		request.Header.Add("X-Real-IP", "118.88.88.88")
		if strings.Contains(requestURI, "http") {
			requestURI = strings.ReplaceAll(requestURI, scheme+request.Host, "")
		}
		remote, err := url.Parse(scheme + proxyDomain + requestURI)
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		needTransport := false
		for _, path := range Path {
			if strings.Contains(request.RequestURI, path) {
				needTransport = true
				break
			}
		}
		if needTransport && request.Method == http.MethodPost {
			proxy.Transport = &capturedTransport{}
			fmt.Printf("Transport:%s(%s)\n", remote, request.Host)
		} else {
			proxy.Transport = &http.Transport{
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
				TLSClientConfig:
				&tls.Config{ServerName: request.Host},
			}
			fmt.Printf("Direct:%s(%s)\n", remote, request.Host)
		}
		proxy.ServeHTTP(resp, request)
	} else {
		scheme := "http://"
		if request.TLS != nil {
			scheme = "https://"
		}
		requestURI := request.RequestURI
		request.Header.Del("x-napm-retry")
		request.Header.Add("X-Real-IP", "118.88.88.88")
		if !strings.Contains(requestURI, "http") {
			requestURI = scheme + request.Host + requestURI
		}
		request.Header.Del("x-napm-retry")
		request.Header.Add("X-Real-IP", "118.88.88.88")
		remote, err := url.Parse(requestURI)
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		fmt.Printf("Direct:%s\n", remote)
		proxy.ServeHTTP(resp, request)
	}
}

type capturedTransport struct {
	// Uncomment this if you want to capture the transport
	// CapturedTransport http.RoundTripper
}

func (t *capturedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	netease := processor.DecodeRequestBody(request)
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		fmt.Println(err)
		return response, err
	}
	processor.DecodeResponseBody(response, netease)
	return response, err
}

func startTlsServer(addr, certFile, keyFile string, handler http.Handler) {
	fmt.Printf("starting TLS Server  %s\n", addr)
	currentPath, error := utils.GetCurrentPath()
	if error != nil {
		fmt.Println(error)
		currentPath = ""
	}
	//fmt.Println(currentPath)
	certFile, _ = filepath.Abs(currentPath + certFile)
	keyFile, _ = filepath.Abs(currentPath + keyFile)
	s := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		panic(err)
	}
}
func startServer(addr string, handler http.Handler) {
	fmt.Printf("starting Server  %s\n", addr)
	s := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
