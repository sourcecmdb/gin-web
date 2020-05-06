package gin_web

import (
	"fmt"
	"github.com/sourcecmdb/gin-web/internal/bytesconv"
	"github.com/sourcecmdb/gin-web/render"
	"github.com/sourcecmdb/gin-web/utils"
	"html/template"
	"net"
	"net/http"
	"os"
	"path"
	"sync"
)

const defaultMultipartMemory = 32 << 20 // 32 m  内存
var (
	default404Body  = []byte("404 page not found")
	default405Body  = []byte("405 method not allowed")
	defaultAppEngin bool
)

// HandlerFunc 将gin-web中间件使用的处理程序定义为返回值
type HandlerFunc func(*Context)

//HandlersChain 定义一个HandlerFunc数组
type HandlersChain []HandlerFunc

// last 返回链中的最后与i个处理程序，即 最后一个处理程序是主要
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// Routeelnfo 表示请求路由的规范 其中包含方法和路径及其处理程序
type RouteInfo struct {
	Method      string      // 方法
	Path        string      // 路由路径
	Handler     string      // Handler 请求头
	HandlerFunc HandlerFunc //
}

// Routesinfo 定义一个Routeinfo数组
type Routesinfo []RouteInfo

// Engine  是框架的实例，它包含多路复用器，中间件和配置设置
// Default  使用New或者Default 创建Engine的实例
type Engine struct {
	RouterGroup
	//如果当前路由无法匹配，但启用了自动重定向，则启用  Enables automatic redirection if the current route can't be matched but a
	//存在带有（不带）斜杠的路径的处理程序。  handler for the path with (without) the trailing slash exists.
	//例如，如果请求/ foo /，但仅存在/ foo的路由，则 For example if /foo/ is requested but a route only exists for /foo, the
	//客户端使用http状态代码301重定向到/ foo进行GET请求 client is redirected to /foo with http status code 301 for GET requests
	//和307（表示所有其他请求方法）。 and 307 for all other request methods.

	RedirectTrailignslash bool

	//如果启用，则路由器尝试修复当前请求路径，如果没有  If enabled, the router tries to fix the current request path, if no
	//为它注册了句柄。  handle is registered for it.
	//首先删除多余的路径元素，例如../或//。   First superfluous path elements like ../ or // are removed
	//之后，路由器对清除的路径进行不区分大小写的查找。   Afterwards the router does a case-insensitive lookup of the cleaned path.
	//如果可以找到此路由的句柄，则路由器进行重定向  If a handle can be found for this route, the router makes a redirection
	//到状态为301的GET请求和307的更正路径   to the corrected path with status code 301 for  GET requests and 307 for
	//所有其他请求方法。   all other request methods.
	//例如，/ FOO和/..//Foo可以重定向到/ foo。   For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash与该选项无关。  RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	//如果启用，路由器会检查是否允许使用其他方法   If enabled, the router checks if another method is allowed for the
	//当前路由，如果当前请求不能被路由。   current route, if the current request can not be routed.
	//如果是这种情况，则使用“不允许的方法”回答请求   If this is the case, the request is answered with 'Method Not Allowed'
	//和HTTP状态代码405。   and HTTP status code 405.
	//如果不允许其他方法，则将请求委托给NotFound   If no other Method is allowed, the request is delegated to the NotFound
	//处理程序。   handler.
	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool

	//＃726＃755如果启用，它将以 #726 #755 If enabled, it will thrust some headers starting with
	//'X-AppEngine ...'，以更好地与该PaaS集成。 'X-AppEngine...' for better integration with that PaaS.
	AppEngine bool

	//如果启用，则将使用url.RawPath查找参数。  If enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	//如果为true，则路径值将不转义。  If true, the path value will be unescaped.
	//如果UseRawPath为false（默认情况下），则UnescapePathValues实际上为true，  If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	//作为url.Path将被使用，它已经被转义了。  as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	//赋予http.Request的ParseMultipartForm的'maxMemory'参数的值 Value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	//方法调用。method call.
	MaxMultipartMemory int64

	//即使有多余的斜杠，也可以从URL解析参数RemoveExtraSlash。 RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	//参见PR＃1817并发布＃1644 See the PR #1817 and issue #1644
	RemoveExtraSlash bool

	delims           render.Delims
	secureJsonPrefix string
	HTMLRender       render.HTMLRender
	FuncMap          template.FuncMap
	allNoRoute       HandlersChain
	allNoMethod      HandlersChain
	noRoute          HandlersChain
	pool             sync.Pool
	trees            methodTrees
}

var _ IRouter = &Engine{}

// New返回一个新的空白Engine实例，不附加任何中间件。 New returns a new blank Engine instance without any middleware attached.
//默认情况下，配置为： By default the configuration is:
// - RedirectTrailingSlash:  true
// - RedirectFixedPath:      false
// - HandleMethodNotAllowed: false
// - ForwardedByClientIP:    true
// - UseRawPath:             false
// - UnescapePathValues:     true
func New() *Engine {
	debugPrintWARWINGNew()
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		FuncMap:                template.FuncMap{},
		RedirectTrailignslash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		ForwardedByClientIP:    true,
		AppEngine:              defaultAppEngin,
		UseRawPath:             false,
		RemoveExtraSlash:       false,
		UnescapePathValues:     true,
		MaxMultipartMemory:     defaultMultipartMemory,
		trees:                  make(methodTrees, 0, 9),
		delims:                 render.Delims{Left: "{{", Right: "}}"},
		secureJsonPrefix:       "while(1)",
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine
}
func (engine *Engine) Use(middleware ...HandlerFunc) IRoutes {

}

// 默认值返回已连接Logger和Recovery中间件的Engine实例。Default returns an Engine instance with the Logger and Recovery middleware already attached.=
func Default() *Engine {
	debugPrintWARWINGDefault()
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine

}
func (engine *Engine) allocateContext() *Context {
	return &Context{engine: engine, KeysMutex: &sync.RWMutex{}}
}

//Delims左右设置模板delims，并返回Engine实例。 // Delims sets template left and right delims and returns a Engine instance.
func (engine *Engine) Delims(left, right string) *Engine {
	engine.delims = render.Delims{Left: left, Right: right}
	return engine
}

//SecureJsonPrefix设置Context.SecureJSON中使用的secureJsonPrefix。 SecureJsonPrefix sets the secureJsonPrefix used in Context.SecureJSON.
func (engine *Engine) SecureJsonPrefix(prefix string) *Engine {
	engine.secureJsonPrefix = prefix
	return engine
}

// SetHTMLTemplate将模板与HTML渲染器关联。 // SetHTMLTemplate associate a template with HTML renderer.
func (engine *Engine) SetHTMLTemplate(templ *template.Template) {
	if len(engine.trees) > 0 {
		debugPrintWARNIGSetHTMLTemplate()
	}
	engine.HTMLRender = render.HTMLProduction{Template: templ.Funcs(engine.FuncMap)}
}

// LoadHTMLGlob加载由glob模式标识的HTML文件 // LoadHTMLGlob loads HTML files identified by glob pattern
//并将结果与HTML渲染器关联// and associates the result with HTML renderer.
func (engine *Engine) LoadHTMLGlob(prttern string) {
	left := engine.delims.Left
	right := engine.delims.Right
	templ := template.Must(template.New("").Delims(left, right).Funcs(engine.FuncMap).ParseGlob(prttern))

	if IsDebugging() {
		debugPrintLoadTemplate(templ)
		engine.HTMLRender = render.HTMLDebug{Glob: prttern, FuncMap: engine.FuncMap, Delims: engine.delims}
		return
	}
	engine.SetHTMLTemplate(templ)
}

// LoadHTMLFiles加载HTML文件的一部分 // LoadHTMLFiles loads a slice of HTML files
//并将结果与HTML渲染器关联。 // and associates the result with HTML renderer.
func (engine *Engine) LoadHTMLFiles(files ...string) {
	if IsDebugging() {
		engine.HTMLRender = render.HTMLDebug{Files: files, FuncMap: engine.FuncMap, Delims: engine.delims}
		return
	}
	templ := template.Must(template.New("").Delims(engine.delims.Left, engine.delims.Right).Funcs(engine.FuncMap).ParseFiles(files...))
	engine.SetHTMLTemplate(templ)
}

// SetFuncMap设置用于template.FuncMap的FuncMap。 // SetFuncMap sets the FuncMap used for template.FuncMap.
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.FuncMap = funcMap
}

func (engine *Engine) rebuild404Handlers() {
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
}

func (engine *Engine) rebuild405Handlers() {
	engine.allNoMethod = engine.combineHandlers(engine.allNoMethod)
}

// NoRoute为NoRoute添加处理程序。 默认情况下，它返回404代码。// NoRoute adds handlers for NoRoute. It return a 404 code by default.
func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	engine.noRoute = handlers
	engine.rebuild404Handlers()
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	utils.Assert1(path[0] == '/', "path must begin with '/' ")
	utils.Assert1(method != "", "HTTP method can not be empty")
	utils.Assert1(len(handlers) > 0, "There must be at least one handler")

	debugPrintRoute(method, path, handlers)
	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}

func iterate(path, method string, routes Routesinfo, root *node) Routesinfo {
	path += root.path
	if len(root.handlers) > 0 {
		handlerFunc := root.handlers.Last()
		routes = append(routes, RouteInfo{
			Method:      method,
			Path:        path,
			Handler:     utils.NameOfFunction(handlerFunc),
			HandlerFunc: handlerFunc,
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
}

// 路线返回一部分已注册的路线，其中包括一些有用的信息，例如： Routes returns a slice of registered routes including some userful information such as:
// http方法，路径和处理程序名称。 the http method , path and the handler name.
func (engine *Engine) Routes() (routes Routesinfo) {
	for _, tree := range engine.trees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

//运行将路由器附加到http.Server并开始侦听和处理HTTP请求 // Run attaches the router to a http.Server and starts listening and serving HTTP requests
//这是http.ListenAndSeve（addr，router）的快捷方式 // It is a shortcut for http.ListenAndSeve(addr,router)
//注意：除非发生错误，否则此方法将无限期阻止调用goroutine // Note: this method will block the calling goroutine indefinitely unless an error happens
func (engine *Engine) Run(addr ...string) (err error) {
	defer func() { debugPrintError(err) }()

	address := utils.ResolverAddress(addr)
	DebugPrint("监听并提供HTTP服务 Listening and serving HTTP on %s\n", address)
	err = http.ListenAndServe(address, engine)
	return
}

// RunTLS将路由器附加到http.Server，并开始侦听和处理HTTPS（安全）请求。 // RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
//这是http.ListenAndServeTLS（addr，certFile，keyFile，router）的快捷方式 // It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
//注意：除非发生错误，否则此方法将无限期阻止调用goroutine。 // Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunTLS(addr, certFile, keyFile string) (err error) {
	DebugPrint("监听并提供HTTPS服务 Listening and serving HTTPS on %s\n", addr)
	defer func() { debugPrintError(err) }()
	err = http.ListenAndServeTLS(addr, certFile, keyFile, engine)
	return
}

// RunUnix将路由器附加到http.Server并开始侦听和处理HTTP请求 // RunUnix attaches the router to a http.Server and starts listening and serving HTTP requests
//通过指定的Unix套接字（即文件）。 // through the specified unix socket (ie. a file).
//注意：除非发生错误，否则此方法将无限期阻止调用goroutine。// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunUnix(file string) (err error) {
	DebugPrint("监听并提供HTTP服务 Listening and serving  HTTP on unix:/%s", file)
	defer func() { debugPrintError(err) }()
	listener, err := net.Listen("nuix", file)
	if err != nil {
		return
	}
	defer listener.Close()
	defer os.Remove(file)

	err = http.Serve(listener, engine)
	return
}

// RunListener将路由器附加到http.Server并开始侦听和处理HTTP请求 // RunListener attaches the router to a http.Server and starts listening and serving HTTP requests
//通过指定的net.Listener // through the specified net.Listener
func (engine *Engine) RunListener(listener net.Listener) (err error) {
	DebugPrint("在侦听器上侦听和服务HTTP与address @％s绑定的内容 Listening and serving HTTP on listener what's bind with address@%s", listener.Addr())
	defer func() { debugPrintError(err) }()
	err = http.Serve(listener, engine)
	return
}

// RunFd将路由器附加到http.Server并开始侦听和处理HTTP请求 // RunFd attaches the router to a http.Server and starts listening and serving HTTP requests
//通过指定的文件描述符。  // through the specified file descriptor.
//注意：除非发生错误，否则此方法将无限期阻止调用goroutine。 // Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunFd(fd int) (err error) {
	DebugPrint("Listening and serving HTTP on fd@%d", fd)
	defer func() { debugPrintError(err) }()

	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd@%d", fd))
	listerner, err := net.FileListener(f)
	if err != nil {
		return
	}
	defer listerner.Close()
	err = engine.RunListener(listerner)
	return
}

func redirectRequest(c *Context) {
	req := c.Request
	rPath := req.URL.Path
	rURL := req.URL.String()

	code := http.StatusMovedPermanently //永久重定向，使用GET方法请求 // permanent redirect, request with GET method
	if req.Method != http.MethodGet {
		code = http.StatusTemporaryRedirect
	}
	DebugPrint("redirecting requess %d:%s --> %s", code, rPath, rURL)
	http.Redirect(c.Writer, req, rURL, code)
	c.writermem.WriteHeaderNow()
}

func redirectTrailingSlash(c *Context) {
	req := c.Request
	p := req.URL.Path
	if prefix := path.Clean(c.Request.Header.Get("X-Forwarded-Prefix")); prefix != "." {
		p = prefix + "/" + req.URL.Path
	}
	req.URL.Path = p + "/"
	if length := len(p); length > 1 && p[length-1] == '/' {
		req.URL.Path = p[:length-1]
	}
	redirectRequest(c)
}

func redirectFixedPath(c *Context, root *node, trailingSlash bool) bool {
	req := c.Request
	rPath := req.URL.Path

	if fixedPath, ok := root.findCaseInsensitivePath(cleanPath(rPath), trailingSlash); ok {
		req.URL.Path = bytesconv.BytesToString(fixedPath)
		redirectRequest(c)
		return true
	}
	return false
}

var mimePlain = []string(MIMEPlain)

func serveError(c *Context, code int, defaultMassags []byte) {
	c.writermem.status = code
	c.Next()
	if c.writermem.Written() {
		return
	}
	if c.writermem.Status() == code {
		c.writermem.Header()["Content-Type"] = mimePlain
		_, err := c.Writer.Write(defaultMassags)
		if err != nil {
			DebugPrint("connot write message to writer during serve error: %V", err)
		}
		return
	}
	c.writermem.WriteHeaderNow()
}

func (engine *Engine) handleHTTPRequest(c *Context) {
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path
	unescape := false
	if engine.UseRawPath && len(c.Request.URL.RawPath) > 0 {
		rPath = c.Request.URL.RawPath
		unescape = engine.UnescapePathValues
	}
	if engine.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}
	//查找给定HTTP方法的reee的根 // find root of the reee for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}

		root := t[i].root
		//在树中查找路线 // Find route in  tree
		value := root.getValue(rPath, c.Params, unescape)
		if value.handlers != nil {
			c.handlers = value.handlers
			c.Params = value.params
			c.fullPath = value.fullPath
			c.Next()
			c.writermem.WriteHeaderNow()
			return
		}

		if httpMethod != "CONNECT" && rPath != "/" {
			if value.tsr && engine.RedirectTrailignslash {
				redirectTrailingSlash(c)
				return
			}
			if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
				return
			}
		}
		break
	}
	if engine.HandleMethodNotAllowed {
		for _, tree := range engine.trees {
			if tree.method == httpMethod {
				continue
			}
			if value := tree.root.getValue(rPath, nil, unescape); value.handlers != nil {
				c.handlers = engine.allNoMethod
				serveError(c, http.StatusMethodNotAllowed, default405Body)
				return
			}
		}
	}
	c.handlers = engine.allNoRoute
	serveError(c, http.StatusNotFound, default404Body)
}

// ServeHTTP符合http.Handler接口。 // ServeHTTP conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.writermem.reset(w)
	c.Request = req
	c.reset()
	engine.handleHTTPRequest(c)
}
