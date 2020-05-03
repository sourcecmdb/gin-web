package gin_web

import (
	"fmt"
	"html/template"
	"runtime"
	"strconv"
	"strings"
)

const gin_webSupporMinGoVer = 0

func IsDebugging() bool {
	return ginMode == debugCode
}
func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[GIN-debug]"+format, values...)
	}
}

func getMinVer(v string) (uint64, error) {
	first := strings.IndexByte(v, '.')
	last := strings.IndexByte(v, '.')
	if first == last {
		return strconv.ParseUint(v[first+1:], 10, 64)
	}
	return strconv.ParseUint(v[first+1:last], 10, 64)
}

func debugPrintWARWINGNew() {
	debugPrint(`[警告]以“调试”模式运行。 在生产中切换到“发布”模式。[WARNING] Running in "debug" mode. Switch to "release" mode in production.
  -使用环境：export GIN_MODE = release - using env:	export GIN_MODE=release
   -使用代码：gin.SetMode（gin.ReleaseMode） - using code:	gin.SetMode(gin.ReleaseMode)
`)
}

func debugPrintWARWINGDefault() {
	if v, e := getMinVer(runtime.Version()); e == nil && v <= gin_webSupporMinGoVer {
		debugPrint(` [警告]现在，杜松子酒需要Go 1.11或更高版本，而不久之后将需要Go 1.12 [WARNING] Now Gin requires Go 1.11 or later and Go 1.12 will be required soon`)
	}
	debugPrint(`[警告]使用已连接Logger和Recovery中间件创建Engine实例  [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached`)
}

func debugPrintLoadTemplate(tmpl *template.Template){
	is IsDebugging(){
		var buf strings.Builder
		for _,tmpl := range tmpl.Templates(){
			buf.WriteString("\t- ")
			buf.WriteString(tmpl.Name())
			buf.WriteString("\n")
		}
		debugPrint("Loaded HTML Templates (%d):\n%s\n",len(tmpl.Templates()),buf.String())
	}
}

func debugPrintWARNIGSetHTMLTemplate(){
	debugPrint(`
`)
}