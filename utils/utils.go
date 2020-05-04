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
			gin_web.DebugPrint("Eenvironment variable PORT=\"%s\"", port)
			return ":" + port
		}

	}
}
