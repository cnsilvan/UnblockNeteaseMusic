package config

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
)

var (
	Port   = flag.String("p", "80:443", "specify server port,such as : \"80:443\"")
	Source = flag.String("o", "kuwo:kugou", "specify server source,such as : \"kuwo:kugou\"")
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

	fmt.Println("--------------------Port------------------------")
	ports := strings.Split(*Port, ":")
	if len(ports) < 1 {
		fmt.Printf("port param invalid: %v \n", *Port)
		return false
	}
	for _, p := range ports {
		fmt.Println(p)
		if m, _ := regexp.MatchString("^\\d+$", p); !m {
			fmt.Printf("port param invalid: %v \n", *Port)
			return false
		}
	}
	fmt.Println("--------------------Source------------------------")
	sources := strings.Split(*Source, ":")
	if len(sources) < 1 {
		fmt.Printf("source param invalid: %v \n", *Source)
		return false
	}
	for _, p := range sources {
		fmt.Println(p)
	}
	return true
}
