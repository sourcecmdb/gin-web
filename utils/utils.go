package utils

import (
	gin_web "github.com/sourcecmdb/gin-web"
	"os"
	"reflect"
	"runtime"
)

func Assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func NameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func ResolverAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			gin_web.DebugPrint("环境变量PORT Eenvironment variable PORT=\"%s\"", port)
			return ":" + port
		}
		gin_web.DebugPrint("环境变量PORT未定义。 默认使用端口：8080 Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("参数太多 too many parameters")
	}
}
