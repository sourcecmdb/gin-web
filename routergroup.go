package gin_web

import "net/http"

//RouterGroup在内部用于配置路由器，RouterGroup与  RouterGroup is used internally to configure router, a RouterGroup is associated with
//前缀和处理程序数组（中间件）。  a prefix and an array of handlers (middleware).
type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

// IRoutes定义了所有路由器句柄接口。 IRoutes defines all router handle interface.
type IRoutes interface {
	Use(...HandlerFunc) IRoutes
	Handle(string, string, HandlerFunc) IRoutes
	Any(string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PATCH(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
	OPTIONS(string, ...HandlerFunc) IRoutes
	HEAD(string, ...HandlerFunc) IRoutes

	StaticFile(string, string) IRoutes
	Static(string, string) IRoutes
	StaticFS(string, http.FileSystem) IRoutes
}

//  IRouter定义了所有路由器句柄接口，包括单路由器和组路由器。 IRouter defines all router handle interface includes single and group router.
type IRouter interface {
	IRoutes
	Group(string, ...http.HandlerFunc) *RouterGroup
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain{
	finalSzie := len(group.Handlers) +len(handlers)
	if finalSzie >= int(abortIndex){
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain,finalSzie)
	copy(mergedHandlers,group.Handlers)
	copy(mergedHandlers[len(group.Handlers):],handlers)
	return mergedHandlers
}