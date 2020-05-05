package gin_web

import (
	"io"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

type ResponesWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// 返回当前请求的HTTP响应状态代码
	Status() int

	//返回已经写入响应 HTTP 主体的字节数
	//  See written()
	Size() int

	//	将字符串写入响应体主体
	WriteString(string) (int, error)
	// 如果响应主体已经写入则返回True
	Written() bool
	// 强制切入http标头（状态代码 响应头）
	WriteHeaderNow()
	// 获取用于服务器推送的 http.Pusher
	Pusher() http.Pusher
}
type responesWriter struct {
	http.ResponseWriter
	size   int
	status int
}

var _ ResponesWriter = &responesWriter{}

func (w *responesWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}
