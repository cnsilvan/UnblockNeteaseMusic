package main

import (
	"config"
	"fmt"
	"host"
	"proxy"
	"version"
)

func main() {
	fmt.Printf(`
		##       ##         ##        ##   ##       ##       ## ##     ## ## ##      ## ## 
		##       ##       ## ##     ## ##  ##       ##    ##      ##      ##      ##      ##
		##       ##      ##  ##    ##  ##  ##       ##   ##               ##     ##
		##       ##     ##   ##   ##   ##  ##       ##    ## ## ##        ##     ##
		##       ##    ##    ##  ##    ##  ##       ##            ##      ##     ## 
		##       ##   ##     ## ##     ##  ##       ##  ##        ##      ##      ##      ##
		## ## ## ##  ##      ####      ##  ## ## ## ##   ## ## ##      ## ## ##    ## ## ##
		
                       %s`+"  by cnsilvan（https://github.com/cnsilvan/UnblockNeteaseMusic） \n", version.Version)
	fmt.Println("--------------------Version--------------------")
	fmt.Println(version.FullVersion())
	if config.ValidParams() {
		fmt.Println("--------------------Config--------------------")
		fmt.Println("port=", *config.Port)
		fmt.Println("source=", *config.Source)
		if host.InitHosts() == nil {
			proxy.InitProxy()
		}
	}
}
