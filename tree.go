package gin_web

import "strings"

// Param 参数是单个URL参数，由键和值组成
type Param struct {
	Key   string
	Value string
}

// 参数是路由器返回的参数切片
// 切片是有序的，第一个URL参数也是第一个切片值
// 因此，通过索引读取值是安全的
type Params []Param
type nodeType uint8

const (
	static nodeType = iota
	root
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	children  []*node
	handlers  HandlersChain
	priority  uint32
	nType     nodeType
	maxParams uint8
	wildChild bool
	fullPath  string
}
type methodTree struct {
	method string
	root   *node
}
type methodTrees []methodTree

// Get返回与给定名称匹配的第一个Param的值。 // Get returns the value of the first Param which key matches the given name.
//如果找不到匹配的Param，则返回一个空字符串。 // If no matching Param is found, an empty string is returned.
func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

func counParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] == ':' || path[i] == '*' {
			n++
		}
	}
	if n > 255 {
		return 255
	}
	return uint8(n)
}

//搜索通配符段，并检查名称中是否包含无效字符。// Search for a wildcard segment and check the name for invalid characters.
//如果未找到通配符war，则返回-1作为索引。// Returns -1 as index, if no wildcard war found.
func findWildcard(path string) (wildcard string, i int, valid bool) {
	// 寻找开始 Find start
	for start, c := range []byte(path) {
		// 通配符以'：'（参数）或'*'（全包）开头 A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}
		//查找结尾并检查无效字符 // Find end and check for invalid characters
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

func (n *node) insertChild(numParams uint8, path string, fullPath string, handlers HandlersChain) {
	for numParams > 0 {
		// 查找前缀，直到第一个通配符  Find prefix until first wildcard
		wildcard, i, valid := findWildcard(path)
		if i < 0 { ////未找到通配符  No wildcard found
			break
		}

		//通配符名称不得包含“：”和“ *” // The wildcard name must not contain ':' and '*'
		if !valid {
			panic("每个路径段只允许使用一个通配符，其具有：' only one wildcard per path segment is allowed, has: '" +
				wildcard + "' 在路上 in path '" + fullPath + "")
		}
		//检查通配符是否具有名称 // check if the wildcard has a name
		if len(wildcard) < 2 {
			panic("通配符必须在路径中使用非空名称命名 wildcards must be named with a non-empty name in path " + fullPath + "'")
		}
		// 检查此节点是否有现有子节点 check if this node has existing chilkdren which would be //
		//如果在此处插入通配符，将无法访问  unreachable if we insert the wildcard here
		if len(n.children) > 0 {
			panic("通配符段 wildcard segment '" + wildcard + "'与现有儿童冲突  conflicts with existing children in path '" + fullPath + "'")
		}

		if wildcard[0] == ':' {
			if i > 0 { //停止  param
				// 在当前通配符之前插入前缀 Insert prefix before the current wildcard
				n.path = path[:i]
				path = path[i:]
			}

			n.wildChild = true
			child := &node{
				nType:     param,
				path:      wildcard,
				maxParams: numParams,
				fullPath:  fullPath,
			}
			n.children = []*node{child}
			n = child
			n.priority++
			numParams--
			//如果路径不是以通配符结尾，则存在	// if the path doesn't end with the wildcard, then there
			//将是另一个以'/'开头的非通配符子路径	// will be another non-wildcard subpath starting with '/'
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]

				child := &node{
					maxParams: numParams,
					priority:  1,
					fullPath:  fullPath,
				}
				n.children = []*node{child}
				n = child
				continue
			}

			//否则，我们完成了将句柄插入新叶子的操作 // Otherwise we're done Insert the handle in the new leaf
			n.handlers = handlers
			return
		}
		// catchAll
		if i+len(wildcard) != len(path) || numParams > 1 {
			panic("包罗万象的路线仅在path中的和处被允许 catch-all route are only allowed at the and of the path in path ' " + fullPath + "")
		}

		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
			panic("与路径中根路径段的现有句柄的全部捕获冲突 catch-all conflicts with existing handle for the path segment root in path ' " + fullPath + "")
		}

		//  //当前固定宽度为'/' /currently fixed width 1 for '/'
		i--
		if path[i] != '/' {
			panic("否/全面通行之前 no / before catch-all in path '" + fullPath + "'")
		}

		n.path = path[:i]

		//第一个节点：具有空路径的catchAll节点	// First node : catchAll node with empty path
		child := &node{
			wildChild: true,
			nType:     catchAll,
			maxParams: 1,
			fullPath:  fullPath,
		}
		//第二个节点：保存变量的节点	// second node: node holding the variable
		child = &node{
			path:      path[i:],
			nType:     catchAll,
			maxParams: 1,
			handlers:  handlers,
			priority:  1,
			fullPath:  fullPath,
		}

		n.children = []*node{child}
		return
	}
	// 如果未找到通配符，只需插入路径和句柄 //If no wildcard was found, simply insert the path and handle
	n.path = path
	n.handlers = handlers
	n.fullPath = fullPath
}

// nodeValue保存（* node）.getValue方法的返回值 //nodeValue holds return values of (*node).getValue method

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longesetCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

//增加给定孩子的优先级，并在必要时重新排序   increments priority of the given child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	//调整位置（移到前面）	// adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {

	}
}

// addRoute将具有给定句柄的节点添加到路径。 // addRoute adds a node with the given handle to the path.
//不是并发安全的！ // Not concurrency-safe!
func (n *node) addRoute(path string, handlers HandlersChain) {
	fullPath := path
	n.priority++
	numParams := counParams(path)
	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(numParams, path, fullPath, handlers)
		n.nType = root
		return
	}
	parentFullPathIndex := 0
walk:
	for {
		//更新当前节点的maxParams // Update maxParams of the current node
		if numParams > n.maxParams {
			n.maxParams = numParams
		}
		//找到最长的公共前缀。	// Find the logest common prefix.
		//这也意味着公共前缀不包含'：'或'*'	// This also implies that the common prefix contains no ':' or '*'
		//，因为现有键不能包含这些字符。	// since the existing key can't contain those chars.
		i := longesetCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := node{
				path:      n.path[i:],
				wildChild: n.wildChild,
				indices:   n.indices,
				children:  n.children,
				handlers:  n.handlers,
				priority:  n.priority - 1,
				fullPath:  n.fullPath,
			}

			// 更新maxParams（所有孩子的最大值）// update maxParams (max of all children)
			for _, v := range child.children {
				if v.maxParams > child.maxParams {
					child.maxParams = v.maxParams
				}
			}

			n.children = []*node{&child}
			//[] byte用于正确的Unicode字符转换，请参见＃65 // []byte for proper unicode char conversion , see #65
			n.indices = string([]byte{n.path[i]})
			n.path = path[i:]
			n.handlers = nil
			n.wildChild = false
			n.fullPath = n.fullPath[:parentFullPathIndex+1]
		}

		//使新节点成为该节点的子节点	// make new node a child of this node
		if i < len(path) {
			path = path[i:]

			if n.wildChild {
				parentFullPathIndex += len(n.path)
				n = n.children[0]
				n.priority++

				//更新子节点的maxParams	// update maxParams of the child node
				if numParams > n.maxParams {
					n.maxParams = numParams
				}
				numParams--

				//检查通配符是否匹配	// check if the wildcard matches
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
					//检查记录器通配符，例如 ：name和：names	// check ofr loger wildcard e.g. :name and :names
					if len(n.path) >= len(path) || path[len(n.path)] == '/' {
						continue walk
					}
				}

				pathSeg := path
				if n.nType != catchAll {
					pathSeg = strings.SplitN(path, "/", 2)[0]
				}
				prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
				panic("'" + pathSeg +
					"' 在新的道路上 in new path '" + fullPath +
					"' 与现有通配符冲突 conflicts with exsting wildcard '" + n.path +
					"' 在灭绝之前 in exsting prefix '" + prefix +
					"'")
			}

			c := path[0]
			//参数后的斜杠	//slash after param
			if n.nType == param && c == '/' && len(n.children) == 1 {
				parentFullPathIndex += len(n.path)
				n = n.children[0]
				n.priority++
				continue walk
			}

			//检查是否存在下一个路径字节的孩子	// check if a child with the next path byte exisit
			for i, max := 0, len(n.indices); i < max; i++ {
				if c == n.indices[i] {
					parentFullPathIndex += len(n.path)
					i = n.incrementChildPrio(i)

				}
			}
		}

	}
}
