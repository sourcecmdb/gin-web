package render

import "net/http"

//渲染接口将通过JSON，XML，HTML，YAML等实现。 Render interface is to be implemented by JSON, XML, HTML, YAML and so on.
type Render interface {
	//渲染使用自定义ContentType写入数据。 // Render writes data with custom ContentType.
	Render(w http.ResponseWriter) error
	// WriteContentType写入自定义ContentType。 // WriteContentType writes custom ContentType.
	WriteContentType(w http.ResponseWriter)
}
