package render

import "text/template"

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
	Instance(string, interface{}) Render
}

// HTMLProduction包含模板参考及其delims。 // HTMLProduction contains template reference and its delims.
type HTMLProduction struct {
	Template *template.Template
	Delims   Delims
}

//HTMLDebug包含模板delims，模式和带有文件列表的功能。 //HTMLDebug contains template delims and pattern and function with file list.
type HTMLDebug struct {
	Files   []string
	Glob    string
	Delims  Delims
	FuncMap template.FuncMap
}

//HTML包含模板引用及其给定接口对象的名称。  HTML contans template reference and its name with given interface object.
type HTML struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

var htmlContentType = []string{"text/html; charset=utf-8"}

// 实例（HTML生产）返回一个实现Render接口的HTML实例。  // Instanace (HTMLProduction) returns an HTML instance which it realizes Render interface.
func (r HTMLProduction) Instanace(name string, data interface{}) Render {
	return HTML{
		Template: r.Template,
		Name:     name,
		Data:     data,
	}
}
