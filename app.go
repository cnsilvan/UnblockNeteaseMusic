package main

import (
	"config"
	"fmt"
	"host"
	"proxy"
	"version"
)

func main() {
	fmt.Println(version.AppVersion())
	//fmt.Println("--------------------Version--------------------")
	//fmt.Println(version.FullVersion())
	if config.ValidParams() {
		fmt.Println("--------------------Config--------------------")
		fmt.Println("port=", *config.Port)
		fmt.Println("source=", *config.Source)
		fmt.Println("certFile=", *config.CertFile)
		fmt.Println("keyFile=", *config.KeyFile)
		fmt.Println("mode=",*config.Mode)
		if host.InitHosts() == nil {
			proxy.InitProxy()
		}
	}
}
