package config

import (
	"flag"
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/utils"
	"github.com/cnsilvan/UnblockNeteaseMusic/version"
	"os"
	"path/filepath"
	"strings"
)

var (
	Port     = flag.Int("p", 80, "specify server port,such as : \"80\"")
	TLSPort  = flag.Int("sp", 443, "specify server tls port,such as : \"443\"")
	Source   = flag.String("o", "kuwo:kugou", "specify server source,such as : \"kuwo:kugou\"")
	CertFile = flag.String("c", "./server.crt", "specify server cert,such as : \"server.crt\"")
	KeyFile  = flag.String("k", "./server.key", "specify server cert key ,such as : \"server.key\"")
	Mode     = flag.Int("m", 1, "specify running mode（1:hosts） ,such as : \"1\"")
	V        = flag.Bool("v", false, "display version info")
)

func ValidParams() bool {
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Println("--------------------Invalid Params------------------------")
		fmt.Printf("Invalid params=%s, num=%d\n", flag.Args(), flag.NArg())
		for i := 0; i < flag.NArg(); i++ {
			fmt.Printf("arg[%d]=%s\n", i, flag.Arg(i))
		}
	}
	if *V {
		fmt.Println(version.FullVersion())
		return false
	}
	sources := strings.Split(strings.ToLower(*Source), ":")
	if len(sources) < 1 {
		fmt.Printf("source param invalid: %v \n", *Source)
		return false
	}
	for _, source := range sources {
		common.Source = append(common.Source, source)
	}

	currentPath, error := utils.GetCurrentPath()
	if error != nil {
		fmt.Println(error)
		currentPath = ""
	}
	//fmt.Println(currentPath)
	certFile, _ := filepath.Abs(*CertFile)
	keyFile, _ := filepath.Abs(*KeyFile)
	_, err := os.Open(certFile)
	if err != nil {
		certFile, _ = filepath.Abs(currentPath + *CertFile)
	}
	_, err = os.Open(keyFile)
	if err != nil {
		keyFile, _ = filepath.Abs(currentPath + *KeyFile)
	}
	*CertFile = certFile
	*KeyFile = keyFile
	return true
}
