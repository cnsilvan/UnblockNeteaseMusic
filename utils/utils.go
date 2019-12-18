package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var JSON = jsoniter.ConfigCompatibleWithStandardLibrary

func FormatMap(data map[string]interface{}) string {
	format := ""
	for key, value := range data {
		format += fmt.Sprintf("%s=%v\n", key, value)
	}
	return format
}
func ReplaceAll(str string, expr string, replaceStr string) string {
	reg := regexp.MustCompile(expr)
	str = reg.ReplaceAllString(str, replaceStr)
	return str
}
func ParseJson(data []byte) map[string]interface{} {
	var result map[string]interface{}
	d := JSON.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	d.Decode(&result)
	return result
}
func ToJson(object interface{}) string {
	json, err := JSON.Marshal(object)
	if err != nil {
		fmt.Println("ToJson Error：", err)
		return "{}"
	}
	return string(json)
}
func Exists(keys []string, h map[string]interface{}) bool {
	for _, key := range keys {
		if !Exist(key, h) {
			return false
		}
	}
	return true
}
func Exist(key string, h map[string]interface{}) bool {
	_, ok := h[key]
	return ok
}
func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
