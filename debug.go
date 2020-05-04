package gin_web

import (
	"fmt"
	"github.com/sourcecmdb/gin-web/utils"
	"html/template"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const gin_webSupporMinGoVer = 0

//DefaultErrorWriter是GIn-Web用于调试错误的默认io.Writer  DefaultErrorWriter is the default io.Writer used by GIn-web to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

func IsDebugging() bool {
	return ginMode == debugCode
}
func DebugPrint(format string, values ...interface{}) {
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
	DebugPrint(`[警告]以“调试”模式运行。 在生产中切换到“发布”模式。[WARNING] Running in "debug" mode. Switch to "release" mode in production.
  -使用环境：export GIN_MODE = release - using env:	export GIN_MODE=release
   -使用代码：gin.SetMode（gin.ReleaseMode） - using code:	gin.SetMode(gin.ReleaseMode)
`)
}

func debugPrintWARWINGDefault() {
	if v, e := getMinVer(runtime.Version()); e == nil && v <= gin_webSupporMinGoVer {
		DebugPrint(` [警告]现在，杜松子酒需要Go 1.11或更高版本，而不久之后将需要Go 1.12 [WARNING] Now Gin requires Go 1.11 or later and Go 1.12 will be required soon`)
	}
	DebugPrint(`[警告]使用已连接Logger和Recovery中间件创建Engine实例  [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached`)
}

func debugPrintLoadTemplate(tmpl *template.Template) {
	IsDebugging()
	{
		var buf strings.Builder
		for _, tmpl := range tmpl.Templates() {
			buf.WriteString("\t- ")
			buf.WriteString(tmpl.Name())
			buf.WriteString("\n")
		}
		DebugPrint("Loaded HTML Templates (%d):\n%s\n", len(tmpl.Templates()), buf.String())
	}
}

func debugPrintWARNIGSetHTMLTemplate() {
	DebugPrint(`[WARNING] Since SetHTMLTemplate() is NOT thread-safe. It should only be calledat initialization. ie. before any route is registered or the router is listening in a socket:

	router := gin.Default()
	router.SetHTMLTemplate(template) // << good place
[警告]由于SetHTMLTemplate（）不是线程安全的。 它只能被称为在初始化时。 即。 在注册任何路由或路由器在套接字中侦听之前：
路由器：= gin.Default（）
router.SetHTMLTemplate（template）// <<好地方
`)
}

//DebugPrintRouteFunc指示调试日志输出格式。  //DebugPrintRouteFunc indicates debug log output format.
var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)

func debugPrintRoute(httpMethod, absolutePath string, handlers HandlersChain) {
	if IsDebugging() {
		nuHandlers := len(handlers)
		handlerName := utils.NameOfFunction(handlers.Last())
		if DebugPrintRouteFunc == nil {
			DebugPrint("%-6s %-25s --> %s (%d hanlders)\n", httpMethod, absolutePath, handlerName, nuHandlers)
		} else {
			DebugPrintRouteFunc(httpMethod, absolutePath, handlerName, nuHandlers)
		}
	}
}

func debugPrintError(err error) {
	if err != nil {
		if IsDebugging() {
			fmt.Fprintf(DefaultErrorWriter, "[GIN_WEB-debug][ERROR] %v\n", err)
		}
	}
}
