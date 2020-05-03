package utils

import (
	"reflect"
	"runtime"
)

func Assert1(guard bool,text string){
	if !guard{
		panic(text)
	}
}

func NameOfFunction(f interface{}) string{
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
