package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"utils"
)

var (
	Port     = flag.String("p", "80:443", "specify server port,such as : \"80:443\"")
	Source   = flag.String("o", "kuwo:kugou", "specify server source,such as : \"kuwo:kugou\"")
	CertFile = flag.String("c", "./server.crt", "specify server cert,such as : \"server.crt\"")
	KeyFile  = flag.String("k", "./server.key", "specify server cert key ,such as : \"server.key\"")
	Mode=flag.String("m", "1", "specify running mode ,such as : \"server.key\"")
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

	//fmt.Println("--------------------Port------------------------")
	ports := strings.Split(*Port, ":")
	if len(ports) < 1 {
		fmt.Printf("port param invalid: %v \n", *Port)
		return false
	}
	for _, p := range ports {
		//fmt.Println(p)
		if m, _ := regexp.MatchString("^\\d+$", p); !m {
			fmt.Printf("port param invalid: %v \n", *Port)
			return false
		}
	}
	//fmt.Println("--------------------Source------------------------")
	sources := strings.Split(*Source, ":")
	if len(sources) < 1 {
		fmt.Printf("source param invalid: %v \n", *Source)
		return false
	}
	//for _, p := range sources {
	//	fmt.Println(p)
	//}
	currentPath, error := utils.GetCurrentPath()
	if error != nil {
		fmt.Println(error)
		currentPath = ""
	}
	//fmt.Println(currentPath)
	certFile, _ := filepath.Abs(*CertFile)
	keyFile, _ := filepath.Abs(currentPath + *KeyFile)
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
