package main

import (
	"UnblockNeteaseMusic/config"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	//_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据
	"UnblockNeteaseMusic/host"
	"UnblockNeteaseMusic/proxy"
	"UnblockNeteaseMusic/version"
	//_ "net/http/pprof" // 必须，引入 pprof 模块
)

func main() {
	fmt.Println(version.AppVersion())
	//fmt.Println("--------------------Version--------------------")
	//fmt.Println(version.FullVersion())
	if config.ValidParams() {
		fmt.Println("--------------------Config--------------------")
		fmt.Println("port=", *config.Port)
		fmt.Println("tlsPort=", *config.TLSPort)
		fmt.Println("source=", *config.Source)
		fmt.Println("certFile=", *config.CertFile)
		fmt.Println("keyFile=", *config.KeyFile)
		fmt.Println("mode=", *config.Mode)
		if host.InitHosts() == nil {
			//go func() {
				//	//	// terminal: $ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
				//	//	// web:
				//	//	// 1、http://localhost:8081/ui
				//	//	// 2、http://localhost:6060/debug/charts
				//	//	// 3、http://localhost:6060/debug/pprof
				//	//	fmt.Println("启动 6060...")
			//	log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
			//}()
			signalChan := make(chan os.Signal, 1)
			exit := make(chan bool, 1)
			go func() {
				sig := <-signalChan
				fmt.Println("\nreceive signal:", sig)
				fmt.Println("restoreHosts ing...")
				err := host.RestoreHosts()
				if err != nil {
					fmt.Println("restoreHosts error:", err)
				}
				exit <- true
			}()
			signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,syscall.SIGSEGV)
			proxy.InitProxy()
			<-exit
			fmt.Println("exiting UnblockNeteaseMusic")
		}
	}
}
