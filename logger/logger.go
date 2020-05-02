package logger

import (
	gin_web "github.com/sourcecmdb/gin-web"
	"io"
	"net/http"
	"time"
)

type consoleColorModeValue int

const (
	autoColor consoleColorModeValue = iota
	disableColor
	forceColor
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

var consoleColorMode = autoColor

//LogFormatter提供传递给LoggerWithFormatter的格式化程序功能的签名  LogFormatter gives the signature of the formatter function passed to LoggerWithFormatter
type LogFormatter func(param LogFormatterParam) string

//LogFormatterParams是在记录时间到来时将处理任何格式化程序的结构 LogFormatterParams is the structure any formatter will be handed when time to log comes
type LogFormatterParam struct {
	Request *http.Request
	// TimeStamp显示服务器返回响应后的时间。 TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode是HTTP响应代码。 StatusCode is HTTP response code.
	StatusCode int
	//延迟是服务器处理某个请求所花费的时间。 Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// ClientIP等于上下文的ClientIP方法。  // ClientIP equals Context's ClientIP method.
	ClientIP string
	//方法是提供给请求的HTTP方法。// Method is the HTTP method given to the request.
	Method string
	// 路径是客户端请求的路径。  Path is a path the client requests.
	Path string
	// 如果在处理请求时发生错误，则设置ErrorMessage。 // ErrorMessage is set if error has occurred in processing the request.
	ErrorMessge string
	//isTerm显示gin的输出描述符是否指向终端。 isTerm shows whether does gin's output descriptor refers to a terminal.
	isTerm bool
	// BodySize是响应主体的大小 // BodySize is the size of the Response Body
	BodySize int
	// 密钥是在请求上下文中设置的密钥。 //Keys are the keys set on the request's context.
	keys map[string]interface{}
}

// LoggerConfig定义Logger中间件的配置。  LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// 可选的。 默认值为gin-web.defaultLogFormatter Optional. Default value is gin.defaultLogFormatter
	Formatter LogFormatter
	//输出是写入日志的写入器。 // Output is a writer where logs are written.
	// 可选的。 默认值为gin.DefaultWriter。 // Optional. Default value is gin.DefaultWriter.
	Output io.Writer
	// SkipPaths是未写入日志的网址路径数组。	// SkipPaths is a url path array which logs are not written.
	// 可选的。	// Optional.
	SkipPaths []string
}

// IsOutputColor指示是否可以将颜色输出到日志。  IsOutputColor indicates whether can colors be outputted to the log.
func (p *LogFormatterParam) IsOutputColor() bool {
	return consoleColorMode == forceColor || (consoleColorMode == autoColor && p.isTerm)
}

// StatusCodeColor是用于将HTTP状态代码正确记录到终端的ANSI颜色。 StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func (p *LogFormatterParam) StatusColor() string {
	code := p.StatusCode

	switch code {
	case code >= http.StatusOK && code < http.StatusMethodNotAllowed:
		return green

	}
}

// MethodColor是用于将http方法正确记录到终端的ANSI颜色。  // MethodColor is the ANSI color for appropriately logging http method to a terminal.
func (p *LogFormatterParam) MethodColor() string {
	method := p.Method
	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

//  defaultLogFormatter是Logger中间件使用的默认日志格式功能。 defaultLogFormatter is the default log format function Logger middleware uses.
var defaultLogFormatter = func(param LogFormatterParam) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCode
		methodColor = param.MethodColor()
	}
}

// LoggerWithConfig实例具有配置的Logger中间件。 LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf LoggerConfig) gin_web.HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		return defaultLogFormatter
	}
}

//Logger实例是一个Logger中间件，该中间件会将日志写入gin.DefaultWriter。  Logger instances a Logger middleware that will write the logs to gin.DefaultWriter.
//默认情况下gin.DefaultWriter = os.Stdout。 By default gin.DefaultWriter = os.Stdout.
func Logger() gin_web.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}
