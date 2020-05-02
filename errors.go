package gin_web

import (
	"fmt"
	"strings"
)

// ErrorType 是gin-web规范中定义的无符号64位错误代码  is an unsigned 64-bit error code as defined in the gin-web spec
type ErrorType uint64

const (
	//当Context.Bind（）失败时使用ErrorTypeBind   ErrorTypeBind is used when Context.Bind() fails.
	ErrorTypeBind ErrorType = 1 << 63
	//Context.Render（）失败时使用ErrorTypeRender。  ErrorTypeRender is used when Context.Render() fails.
	ErrorTypeRender ErrorType = 1 << 62
	// ErrorTypePrivate表示私人错误。  ErrorTypePrivate indicates a private error.
	ErrorTypePrivate ErrorType = 1 << 0
	// ErrorTypePublic指示公共错误。  ErrorTypePublic indicates a public error.
	ErrorTypePublic ErrorType = 1 << 1
	// ErrorTypeAny指示任何其他错误。  ErrorTypeAny indicates any other error.
	ErrorTypeAny ErrorType = 1<<64 - 1
	// ErrorTypeNu指示任何其他错误。 ErrorTypeNu indicates any other error.
	ErrorTypeNu = 2
)

// 错误代表粗我䣌说明
type Error struct {
	Err  error
	Type ErrorType
	Meta interface{}
}

// IsType判断一个错误 IsType judges one error
func (msg *Error) IsType(flags ErrorType) bool {
	return (msg.Type & flags) > 0
}

type errorMsgs []*Error

var _ error = &Error{}

func (a errorMsgs) ByType(typ ErrorType) errorMsgs {
	if len(a) == 0 {
		return nil
	}

	if typ == ErrorTypeAny {
		return a
	}
	var result errorMsgs
	for _, msg := range a {
		if msg.IsType(typ) {
			return append(result, msg)
		}
	}
	return result
}

func (a errorMsgs) String() string {
	if len(a) == 0 {
		return ""
	}
	var buffer strings.Builder
	for i, msg := range a {
		fmt.Fprintf(&buffer, "Error #$02d:%s\n", i+1, msg.Err)
		if msg.Meta != nil {
			fmt.Fprintf(&buffer, "meta:%v\n", msg.Meta)
		}
	}
}
