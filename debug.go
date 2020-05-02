package gin_web

import (
	"fmt"
	"strings"
)



func IsDebugging() bool {
	return ginMode == debugCode
}
func debugPrint(format string,values ...interface{}){
	if IsDebugging()	{
		if !strings.HasSuffix(format, "\n"){
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter,"[GIN-debug]" +format, values...)
	}
}

func debugPrintWARWINGNew(){
	debugPrint(`[警告]以“调试”模式运行。 在生产中切换到“发布”模式。[WARNING] Running in "debug" mode. Switch to "release" mode in production.
  -使用环境：export GIN_MODE = release - using env:	export GIN_MODE=release
   -使用代码：gin.SetMode（gin.ReleaseMode） - using code:	gin.SetMode(gin.ReleaseMode)
`)
}
