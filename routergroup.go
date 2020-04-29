package gin_web
//RouterGroup在内部用于配置路由器，RouterGroup与  RouterGroup is used internally to configure router, a RouterGroup is associated with
//前缀和处理程序数组（中间件）。  a prefix and an array of handlers (middleware).
type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine *Engine
	root bool
}
