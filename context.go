package gin_web

import (
	"github.com/sourcecmdb/gin-web/binding"
	"github.com/sourcecmdb/gin-web/render"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

//最常见的数据格式的Content-Type MIME // Content-Type MIME of the most common data formats
const (
	MIMEPlain = binding.MIMEPlain
)

const abortIndex int8 = math.MaxInt8 / 2

// 上下文是是最重要的部分。它允许我们在中间件之间的传递变量
// 管理流程 例如验证请求的json并呈现json响应
type Context struct {
	writermem responesWriter
	Request   *http.Request
	Writer    ResponesWriter

	Params   Params
	handlers HandlersChain
	index    int8
	fullPath string

	engine *Engine

	// 此互斥锁保护关键映射 This mutex protect Keys map
	KeysMutex *sync.RWMutex

	// 键是专门针对每个请求的上下文的键值对 Keys is a key/value pair exclusively for the context of each request
	Keys map[string]interface{}

	// 错误是使用此上下文的所有处理程序/中间附件带的错误列表 Errors is a list of errors attached to all the handlers/middlewes who used this context
	Errors errorMsgs

	// 接受定义用于内容协商的手动接受格式的列表。 Accepted defines a list of manually accepted formats for content negotiation.
	Accepted []string

	// queryCache使用url.ParseQuery从c.Request.URL.Query（）缓存了参数查询结果  queryCache use url.ParseQuery cached the param query result from c.Request.URL.Query()
	queryCache url.Values

	//formCache使用url.ParseQuery缓存的PostForm包含从POST，PATCH， formCache use url.ParseQuery cached PostForm contains the parsed form data from POST, PATCH,
	// 或PUT主体参数。  or PUT body parameters.
	formCache url.Values

	// SameSite允许服务器定义cookie属性，从而无法   SameSite allows a server to define a cookie attribute making it impossible for
	// 浏览器将此Cookie与跨站点请求一起发送。  the browser to send this cookie along with cross-site requests.
	sameSite http.SameSite
}

func (c *Context) requestHeader(key string) string {
	return c.Request.Header.Get(key)
}
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) ClientIP() string {
	if c.engine.ForwardedByClientIP {
		clientIP := c.requestHeader("X-Formwarded-For")
		clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
		if clientIP == "" {
			clientIP = strings.TrimSpace(c.requestHeader("X-Real-Ip"))
		}
		if clientIP != "" {
			return clientIP
		}
	}
	if c.engine.AppEngine {
		if addr := c.requestHeader("X-Appenigne-Remote-Addr"); addr != "" {
			return addr
		}
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

//错误将错误附加到当前上下文。 错误被推送到错误列表。 // Error attaches an error to the current context. The error is pushed to a list of errors.
//对于解析请求期间发生的每个错误，最好都调用Error。 // It's a good idea to call Error for each error that occurred during the resolution of a request.
//中间件可用于收集所有错误并将它们一起推送到数据库中， // A middleware can be used to collect all the errors and push them to a database together,
//打印日志，或将其附加到HTTP响应中。 // print a log, or append it in the HTTP response.
//如果err为nil，错误将惊慌。 // Error will panic if err is nil.

func (c *Context) Error(err error) *Error {
	if err == nil {
		panic("err is nil")
	}
	parsedError, ok := err.(*Error)
	if !ok {
		parsedError = &Error{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}
	c.Errors = append(c.Errors, parsedError)
	return parsedError
}

//状态设置HTTP响应代码 //Status sets the HTTP response code
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// AbortWithStatusJSON内部调用`Abort（）`，然后调用`JSON // AbortWithStatusJSON calls `Abort()` and then `JSON` internally.
//此方法停止链，写入状态代码并返回JSON正文。 // This method stops the chain, writes the status code and return a JSON body.
//还将Content-Type设置为“ application / json”。// It also sets the Content-Type as "application/json".
func (c *Context) Abort() {
	c.index = abortIndex
}

// AbortWithError内部调用`AbortWithStatus（）`和`Error（）`。 // AbortWithError calls `AbortWithStatus()` and `Error()` internally.
//此方法停止链，写入状态代码，并将指定的错误推送到`c.Errors`。 // This method stops the chain, writes the status code and pushes the specified error to `c.Errors`.
//有关更多详细信息，请参见Context.Error（）。 // See Context.Error() for more details.
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
}

func (c *Context) reset() {
	c.Writer = &c.writermem
	c.Params = c.Params[0:0]
	c.handlers = nil
	c.index = -1
	c.KeysMutex = &sync.RWMutex{}
	c.fullPath = ""
	c.Keys = nil
	c.Errors = c.Errors[0:0]
	c.Accepted = nil
	c.queryCache = nil
	c.formCache = nil
}

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}

// Render writes the response headers and calls render .Render to render data.
func (c *Context) Render(code int, r render.Render) {
	c.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Writer)
		c.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Writer); err != nil {
		panic(err)
	}
}

// HTML呈现由其文件名指定的HTTP模板。 // HTML renders the HTTP template specified by its file name.
//它还会更新HTTP代码，并将Content-Type设置为“ text / html // It also updates the HTTP code and sets the Content-Type as "text/html".
// See http://golang.org/doc/articles/wiki/
func (c *Context) HTML(code int, name string, obj interface{}) {
	instance := c.engine.HTMLRender.Instance(name, obj)
	c.Render(code, instance)
}
