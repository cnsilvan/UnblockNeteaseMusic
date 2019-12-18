package host

import (
	"bufio"
	"UnblockNeteaseMusic/config"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ProxyIp     = "127.0.0.1"
	ProxyDomain = map[string]string{
		"music.163.com":            "59.111.181.35",
		"interface.music.163.com":  "59.111.181.35",
		"interface3.music.163.com": "59.111.181.35",
		"apm.music.163.com":        "59.111.181.35",
		"apm3.music.163.com":       "59.111.181.35",
	}
	HostDomain = map[string]string{
		"music.163.com":           "59.111.181.35",
		"interface.music.163.com": "59.111.181.35",
	}
	//ProxyDomain = map[string]string{
	//	"music.163.com":            "59.111.181.35",
	//	"interface.music.163.com":  "59.111.181.35",
	//	"interface3.music.163.com": "59.111.181.35",
	//	"apm.music.163.com":        "59.111.181.35",
	//	"apm3.music.163.com":       "59.111.181.35",
	//	"music.httpdns.c.163.com":  "59.111.181.35",
	//	"httpdns.n.netease.com":    "59.111.179.213",
	//}
)

func getWinSystemDir() string {
	dir := ""
	if runtime.GOOS == "windows" {
		dir = os.Getenv("windir")
	}

	return dir
}
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//if os.IsNotExist(err) {
	//	return false, nil
	//}
	return false, err
}
func appendToFile(fileName string, content string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("appendToFile cacheFileList.yml file create failed. err: " + err.Error())
	} else {
		defer f.Close()
		//n, _ := f.Seek(0, io.SeekEnd)
		//_, err = f.WriteAt([]byte(content), n)
		_, err = f.WriteString(content)
		if err != nil {
			fmt.Println("appendToFile write file fail:", err)
			return err
		}
	}
	return err
}
func restoreHost(hostPath string) error {

	return nil
}
func appendToHost(hostPath string) error {
	content := " \n# UnblockNetEaseMusic（Go）\n"
	for domain, _ := range HostDomain {
		content += ProxyIp + " " + domain + "\n"
	}
	return appendToFile(hostPath, content)
}
func backupHost(hostPath string) (bool, error) {
	containsProxyDomain := false
	host, err := os.Open(hostPath)
	if err != nil {
		fmt.Println("open file fail:", err)
		return containsProxyDomain, err
	}
	defer host.Close()
	gBackup, err := os.Create(hostPath + ".gBackup")
	if err != nil {
		fmt.Println("Open write file fail:", err)
		return containsProxyDomain, err
	}
	defer gBackup.Close()
	br := bufio.NewReader(host)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read err:", err)
			return containsProxyDomain, err
		}
		newLine := string(line)
		if !containsProxyDomain {
			if strings.Contains(strings.ToUpper(newLine), strings.ToUpper("UnblockNetEaseMusic")) {
				containsProxyDomain = true
				fmt.Println("Found UnblockNetEaseMusic Line")
			}
			for domain, _ := range ProxyDomain {
				if strings.Contains(newLine, domain) {
					containsProxyDomain = true
					fmt.Println("Found ProxyDomain Line")
				}
			}
		}
		_, err = gBackup.WriteString(newLine + "\n")
		if err != nil {
			fmt.Println("write to file fail:", err)
			return containsProxyDomain, err
		}
	}
	return containsProxyDomain, nil
}

// Exclude UnblockNetEaseMusic related host
func excludeRelatedHost(hostPath string) error {
	host, err := os.Create(hostPath)
	if err != nil {
		fmt.Println("open file fail:", err)
		return err
	}
	defer host.Close()
	gBackup, err := os.Open(hostPath + ".gBackup")
	if err != nil {
		fmt.Println("Open write file fail:", err)
		return err
	}
	defer gBackup.Close()
	br := bufio.NewReader(gBackup)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read err:", err)
			return err
		}
		newLine := string(line)
		needWrite := true
		for domain, _ := range ProxyDomain {
			if strings.Contains(newLine, domain) {
				needWrite = false
				break
			}
		}
		if needWrite && strings.Contains(strings.ToUpper(newLine), strings.ToUpper("UnblockNetEaseMusic")) {
			needWrite = false
		}
		if needWrite && len(strings.TrimSpace(newLine)) == 0 {
			needWrite = false
		}
		if needWrite {
			_, err = host.WriteString(newLine + "\n")
			if err != nil {
				fmt.Println("write to file fail:", err)
				return err
			}
		}
	}
	return nil
}
func resolveIp(domain string) (ip string, err error) {

	return "", nil
}
func resolveIps() error {
	for domain, _ := range HostDomain {
		rAddr, err := net.ResolveIPAddr("ip", domain)
		if err != nil {
			fmt.Printf("Fail to resolve %s, %s\n", domain, err)
			return err
		}
		if len(rAddr.IP) == 0 {
			fmt.Printf("Fail to resolve %s,IP nil\n", domain)
			return fmt.Errorf("Fail to resolve  %s,Ip length==0 \n", domain)
		}
		HostDomain[domain] = rAddr.IP.String()

	}
	return nil
}
func getHostsPath() (string, error) {
	hostsPath := "/etc/hosts"
	if runtime.GOOS == "windows" {
		hostsPath = getWinSystemDir()
		hostsPath = filepath.Join(hostsPath, "system32", "drivers", "etc", "hosts")
	} else {
		hostsPath = filepath.Join(hostsPath)
	}
	if exist, err := fileExists(hostsPath); !exist {
		fmt.Println("Not Found Host File：", hostsPath)
		return hostsPath, err
	}
	return hostsPath, nil
}
func InitHosts() error {
	fmt.Println("-------------------Init Host-------------------")
	if *config.Mode == 1 { //hosts mode
		hostsPath, err := getHostsPath()
		if err == nil {
			containsProxyDomain := false
			containsProxyDomain, err = backupHost(hostsPath)
			if err == nil {
				if containsProxyDomain {
					if err = excludeRelatedHost(hostsPath); err == nil {
						err = resolveIps()
						if err != nil {
							return err
						}
						fmt.Println("HostDomain:", HostDomain)
					}
				} else {
					err = resolveIps()
					if err != nil {
						return err
					}
					fmt.Println("HostDomain:", HostDomain)
				}
				if err = appendToHost(hostsPath); err == nil {

				}
			}
		}
		return err
	} else {
		err := resolveIps()
		if err != nil {
			return err
		}
		fmt.Println("HostDomain:", HostDomain)
		return err
	}

}
