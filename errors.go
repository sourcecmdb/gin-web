package gin_web

// ErrorType 是gin-web规范中定义的无符号64位错误代码  is an unsigned 64-bit error code as defined in the gin-web spec
type ErrorType uint64

const (
 //当Context.Bind（）失败时使用ErrorTypeBind   ErrorTypeBind is used when Context.Bind() fails.
 ErrorTypeBind ErrorType = 1 << 63
 //Context.Render（）失败时使用ErrorTypeRender。  ErrorTypeRender is used when Context.Render() fails.
 ErrorTypeRender ErrorType = 1<<62
 // ErrorTypePrivate表示私人错误。  ErrorTypePrivate indicates a private error.
 ErrorTypePrivate ErrorType = 1 << 0
 // ErrorTypePublic指示公共错误。  ErrorTypePublic indicates a public error.
 ErrorTypePublic ErrorType = 1 << 1
 // ErrorTypeAny指示任何其他错误。  ErrorTypeAny indicates any other error.
 ErrorTypeAny ErrorType = 1<<64 -1
 // ErrorTypeNu指示任何其他错误。 ErrorTypeNu indicates any other error.
 ErrorTypeNu = 2
)

// 错误代表粗我䣌说明
type Error struct {
	Err error
	Type ErrorType
	Meta interface{}
}

type errorMsgs []*Error

var _ error = &Error{}

