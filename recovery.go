package gin_web

import (
	"bytes"
	"fmt"
	"github.com/sourcecmdb/gin-web/logger"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

var (
	dunno     = []byte("???")
	dot       = []byte(".")
	centerDot = []byte("·")
	slash     = []byte("/")
)

//function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	//名称包括包的路径名，这是不必要的 // The name includes the path name to the package, which is unnecessary
	//，因为已经包含了文件名。 另外，它具有中心点。 // since the file name is already included.  Plus, it has center dots.
	//也就是说， // That is, we see
	//运行/调试。* T·ptrmethod //	runtime/debug.*T·ptrmethod
	//并且想要 // and want
	//* T.ptrmethod //	*T.ptrmethod
	//另外，打包路径中可能包含点（例如code.google.com / ...）， // Also the package path might contains dot (e.g. code.google.com/...),
	//因此，首先消除路径前缀 // so first eliminate the path prefix
	if lasfSlash := bytes.LastIndex(name, slash); lasfSlash >= 0 {
		name = name[lasfSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)

	return name
}

//source返回第n行的空间修剪片段。 source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- //在堆栈跟踪中，行是1索引的，但我们的数组是0索引的 in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines[n]) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// stack返回格式正确的堆栈帧，跳过跳过帧。 stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // 返回的数据 the returned data
	//当我们走路时，我们打开文件并读取它们。 这些变量当前记录 // As we loop, we open files and read them. These variables record the currently
	//已加载的文件。 // loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { //跳过预期的帧数 // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		//至少打印此内容。 如果我们找不到源，它将不会显示。 Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func timeFormat(t time.Time) string {
	var timeString = t.Format("2006/01/02 - 15:04:05")
	return timeString
}

// RecoveryWithWriter为给定的编写器返回一个中间件，该中间件可以从任何紧急情况中恢复，如果有中间件，则可以写入500。 RecoveryWithWriter returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func RecoveryWithWriter(out io.Writer) HandlerFunc {
	var logger *log.Logger
	if out != nil {
		logger = log.New(out, "\n\n\x1b[31m", log.LstdFlags)
	}
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				//检查连接是否断开，如果不是	// Check for a broken connection, as it is not really a
				//条件，这保证了紧急堆栈跟踪。		// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				if logger != nil {
					stack := stack(3)
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ":*"
						}
					}
					if brokenPipe {
						logger.Printf("%s\n%s%s", err, string(httpRequest), reset)
					} else if IsDebugging() {
						logger.Printf("[Recovery] %s panic recovered:\n%s\n%s\n%s%s", timeFormat(time.Now()), strings.Join(headers, "\r\n"), err, stack, reset)
					} else {
						logger.Printf("[Recovery %s panic recovered:\n%s\n%s%s", timeFormat(time.Now()), err, stack, reset)
					}
				}

				// 如果连接中断，则无法向其写入状态 If the connection is dead we can't write a status to it
				if brokenPipe {
					c.Error(err.(error)) // nolint:errcheck
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}

// 恢复返回的中间件可从任何紧急情况中恢复，如果有中间件，则写入500。 Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() HandlerFunc {
	return RecoveryWithWriter(DefaultWriter)
}
