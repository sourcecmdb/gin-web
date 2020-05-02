package gin_web

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// 上下文是是最重要的部分。它允许我们在中间件之间的传递变量
// 管理流程 例如验证请求的json并呈现json响应
type Context struct {
	writername responesWriter
	Request    *http.Request
	Writer     ResponesWriter

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
