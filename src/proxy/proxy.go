package proxy

import (
	"bytes"
	"config"
	"fmt"
	"host"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"processor"
	"strings"
	"time"
	"version"
)

type HttpHandler struct{}

func InitProxy() {
	fmt.Println("-------------------Init Proxy-------------------")
	go startTlsServer("0.0.0.0:443", *config.CertFile, *config.KeyFile, &HttpHandler{})
	startServer("0.0.0.0:80", &HttpHandler{})
}
func (h *HttpHandler) ServeHTTP(resp http.ResponseWriter, request *http.Request) {
	requestURI := request.RequestURI
	uriBytes := []byte(requestURI)
	left := uriBytes[:(len(uriBytes) / 2)]
	right := uriBytes[len(uriBytes)/2:]
	scheme := "http://"
	if request.TLS != nil {
		scheme = "https://"
	}
	if strings.Contains(request.Host, "localhost") || strings.Contains(request.Host, "127.0.0.1") || strings.Contains(request.Host, "0.0.0.0") || (len(requestURI) > 1 && strings.Count(requestURI, "/") > 1 && bytes.EqualFold(left, right)) {
		//cause infinite loop
		requestURI = scheme + request.Host
		if bytes.EqualFold(left, right) {
			requestURI += string(left)
		} else {
			requestURI += string(uriBytes)
		}
		fmt.Printf("Abandon:%s\n", requestURI)
		resp.WriteHeader(200)
		resp.Write([]byte(version.AppVersion()))
		return
	}
	if proxyDomain, ok := host.ProxyDomain[request.Host]; ok && !strings.Contains(requestURI, "stream") {
		if strings.Contains(requestURI, "http") {
			requestURI = request.URL.Path
		}
		urlString := scheme + proxyDomain + requestURI
		fmt.Printf("Transport:%s(%s)\n", urlString, request.Host)
		netease := processor.RequestBefore(request)
		//fmt.Printf("{path:%s,web:%v,encrypted:%v}\n", netease.Path, netease.Web, netease.Encrypted)
		response, err := processor.Request(request, urlString)
		if err != nil {
			fmt.Println("Request error:", urlString)
			panic(err)
		}
		defer response.Body.Close()
		processor.RequestAfter(request,response, netease)
		for name, values := range response.Header {
			resp.Header()[name] = values
			//fmt.Println(name,"=",values)
		}
		resp.WriteHeader(response.StatusCode)
		_, err = io.Copy(resp, response.Body)
		if err != nil {
			fmt.Println("io.Copy error:", urlString)
			panic(err)
		}
		response.Body.Close()
		//resp.Write(body)
	} else {
		if strings.Contains(requestURI, "http") {
			requestURI = request.URL.Path
		}
		if proxyDomain, ok := host.ProxyDomain[request.Host]; ok {
			requestURI = scheme + proxyDomain + requestURI
		} else {
			requestURI = scheme + request.Host + requestURI
		}

		remote, err := url.Parse(requestURI)
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		fmt.Printf("Direct:%s\n", remote)
		proxy.ServeHTTP(resp, request)
	}
}

func startTlsServer(addr, certFile, keyFile string, handler http.Handler) {
	fmt.Printf("starting TLS Server  %s\n", addr)
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
