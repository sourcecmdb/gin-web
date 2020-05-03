package gin_web

import (
	"github.com/sourcecmdb/gin-web/render"
	"github.com/sourcecmdb/gin-web/utils"
	"html/template"
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
		engine.trees = append(engine.trees, methodTree{method: method,root: root})
	}
	root.addRoute(path,handlers)
}
