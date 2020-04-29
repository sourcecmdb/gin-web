package render
// Delims代表用于HTML模板渲染的一组左右定界符。   Delims represents a set of Left and Right delimiters for HTML template rendering.
type Delims struct {
	//左定界符，默认为{{。  Left delimiter, defaults to {{.
	Left string
	//右定界符，默认为}}。 	Right delimiter, defaults to }}.
	Right string
}
// HTMLRender接口将由HTMLProduction和HTMLDebug实现。	HTMLRender interface is to be implemented by HTMLProduction and HTMLDebug.
type HTMLRender interface {
	////实例返回一个HTML实例。 // Instance returns an HTML instance.
	Instance(string,interface{}) Render
}
