package gin_web

// Internal helper to lazily cereate a buffer if necessary
// Calls to this function get inlined
func bufApp(buf *[]byte, s string, w int, c byte) {
	b := *buf
	if len(b) == 0 {
		//到目前为止，尚未修改原始字符串。	// No modification of the original string so far.
		//如果下一个字符与原始字符串相同，则执行	// If the next character is the same as in the original string, we do
		//不必分配缓冲区。	// not yet have to allocate a buffer.
		if s[w] == c {
			return
		}

		//否则使用堆栈缓冲区（如果足够大），或者	// Otherwise use either the stack buffer, if it is large enough, or
		//在堆上分配一个新的缓冲区，并复制所有以前的字符。	// allocate a new buffer on the heap, and copy all previous characters.
		if l := len(s); l > cap(b) {
			*buf = make([]byte, len(s))
		} else {
			*buf = (*buf)[:l]
		}
		b = *buf

		copy(b, s[:w])
	}
	b[w] = c

}

// cleanPath是path的URL版本.Clean，它返回规范的URL路径 // cleanPath is the URL version of path.Clean, it returns a canonical URL path
//对于p，消除。 和..元素 // for p, eliminating . and .. elements.
//
//重复应用以下规则，直到无法进行进一步处理为止 // The following rules are applied iteratively until no further processing can
// 做完了： // be done:
// 1.用一个斜杠替换多个斜杠。 //	1. Replace multiple slashes with a single slash.
// 2.消除每个。 路径名元素（当前目录）。 //	2. Eliminate each . path name element (the current directory).
// 3.删除每个内部..路径名元素（父目录） //	3. Eliminate each inner .. path name element (the parent directory)
//及其前面的非..元素。 //	   along with the non-.. element that precedes it.
// 4.消除以..元素开头的根路径： //	4. Eliminate .. elements that begin a rooted path:
//，即在路径的开头用“ /”替换“ / ..”。 //	   that is, replace "/.." by "/" at the beginning of a path.
//
//如果此过程的结果为空字符串，则返回“ /”。 // If the result of this process is an empty string, "/" is returned.
func cleanPath(p string) string {
	const stackBufSize = 128
	//将空字符串转换为“ /” 	// Turn empty string into "/"
	if p == "" {
		return "/"
	}
	//在堆栈上合理大小的缓冲区以避免在常见情况下的分配。	// Reasonably sized buffer on stack to avoid allocations in the common case.
	//如果需要更大的缓冲区，则会动态分配它。	// If a larger buffer is required, it gets allocated dynamically.
	buf := make([]byte, 0, stackBufSize)
	n := len(p)
	//不变量：	// Invariants:
	//从路径读取； r是要处理的下一个字节的索引。	//      reading from path; r is index of next byte to process.
	//写入buf; w是要写入的下一个字节的索引。	//      writing to buf; w is index of next byte to write.

	//路径必须以'/'开头	// path must start with '/'
	r := 1
	w := 1

	if p[0] != '/' {
		r = 0

		if n+1 > stackBufSize {
			buf = make([]byte, n+1)
		} else {
			buf = buf[:n+1]
		}
		buf[0] = '/'
	}
	trailing := n > 1 && p[n-1] == '/'

	//如果没有像路径包这样的“ lazybuf”，则笨拙一些，但是循环 // A bit more clunky without a 'lazybuf' like the path package, but the loop
	//完全内联（bufApp调用）。 // gets completely inlined (bufApp calls).
	//循环没有昂贵的函数调用（1x make除外）//因此，与路径包相比，此循环没有昂贵的函数 // loop has no expensive function calls (except 1x make)		// So in contrast to the path package this loop has no expensive function
	//调用（必要时使用make除外）。	// calls (except make, if needed).
	for r < n {
		switch {
		case p[r] == '/':
			//空的路径元素，末尾添加斜杠	// empty path element, trailing slash is added after the end
			r++
		case p[r] == '.' && r+1 == n:
			trailing = true
			r++
		case p[r] == '.' && p[r+1] == '/':
			//。 元件 // . element
			r += 2
		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			//..元素：最后删除/ // .. element: remove to last /
			r += 3
			if w > 1 {
				// can backtrack
				w--
				if len(buf) == 0 {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}
		default:
			//实际路径元素	// Real path element
			//如果需要的话添加斜杠	// Add slash if needed
			if w > 1 {
				bufApp(&buf, p, w, '/')
				w++
			}

			// Copy element
			for r < n && p[r] != '/' {
				bufApp(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}
	//重新添加斜杠	// Re-append trailing slash
	if trailing && w > 1 {
		bufApp(&buf, p, w, '/')
		w++
	}
	//如果原始字符串未修改（或仅在结尾处缩短），	// If the original string was not modified (or only shortened at the end),
	//返回原始字符串的相应子字符串。   // return the respective substring of the original string.
	//否则从缓冲区返回一个新字符串。	// Otherwise return a new string from the buffer.
	if len(buf) == 0 {
		return p[:w]
	}
	return string(buf[:w])
}
