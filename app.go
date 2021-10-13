package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cnsilvan/UnblockNeteaseMusic/config"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider"
	"github.com/cnsilvan/UnblockNeteaseMusic/version"

	//_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据
	//_ "net/http/pprof" // 必须，引入 pprof 模块

	"github.com/cnsilvan/UnblockNeteaseMusic/host"
	"github.com/cnsilvan/UnblockNeteaseMusic/proxy"
)

func main() {
	//log.Println("--------------------Version--------------------")
	//fmt.Println(version.AppVersion())
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recover panic : ", r)
			restoreHosts()
		}
	}()
	if config.ValidParams() {
		log.Println(version.AppVersion())
		log.Println("--------------------Config--------------------")
		log.Println("port=", *config.Port)
		log.Println("tlsPort=", *config.TLSPort)
		log.Println("source=", *config.Source)
		log.Println("certFile=", *config.CertFile)
		log.Println("keyFile=", *config.KeyFile)
		log.Println("logFile=", *config.LogFile)
		log.Println("mode=", *config.Mode)
		log.Println("endPoint=", *config.EndPoint)
		log.Println("forceBestQuality=", *config.ForceBestQuality)
		log.Println("searchLimit=", *config.SearchLimit)
		log.Println("blockUpdate=", *config.BlockUpdate)
		log.Println("BlockAds=", *config.BlockAds)
		log.Println("EnableLocalVip=", *config.EnableLocalVip)
		log.Println("UnlockSoundEffects=", *config.UnlockSoundEffects)
		log.Println("QQCookieFile=", *config.QQCookieFile)
		if host.InitHosts() == nil {
			//go func() {
			//	//	// terminal: $ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
			//	//	// web:
			//	//	// 1、http://localhost:8081/ui
			//	//	// 2、http://localhost:6060/debug/charts
			//	//	// 3、http://localhost:6060/debug/pprof
			//	//	log.Println("start 6060...")
			//	log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
			//}()

			signalChan := make(chan os.Signal, 1)
			exit := make(chan bool, 1)
			go func() {
				sig := <-signalChan
				log.Println("\nreceive signal:", sig)
				restoreHosts()
				exit <- true
			}()
			signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGKILL)
			proxy.InitProxy()
			provider.Init()
			<-exit
			log.Println("exiting UnblockNeteaseMusic")
		}
	} else {
		fmt.Println(version.AppVersion())
	}
}
func restoreHosts() {
	if *config.Mode == 1 {
		log.Println("restoreHosts...")
		err := host.RestoreHosts()
		if err != nil {
			log.Println("restoreHosts error:", err)
		}
	}
}
