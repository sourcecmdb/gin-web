package gin_web

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
		if len(wildcard) <2{
			panic("通配符必须在路径中使用非空名称命名 wildcards must be named with a non-empty name in path "+ fullPath+"'")
		}
		// 检查此节点是否有现有子节点 check if this node has existing chilkdren which would be //
		//如果在此处插入通配符，将无法访问  unreachable if we insert the wildcard here
		if len(n.children)>0{
			panic("通配符段 wildcard segment '" + wildcard + "'与现有儿童冲突  conflicts with existing children in path '" + fullPath + "'")
		}

		if wildcard[0] == ':'{
			if i >0{ //停止  param
				// 在当前通配符之前插入前缀 Insert prefix before the current wildcard
				n.path = path[:i]
				path = path[i:]

				n.wildChild = true
				child := &node{
					nType :param,
					path:wildcard,
				}
			}
		}
	}
}

// addRoute将具有给定句柄的节点添加到路径。 // addRoute adds a node with the given handle to the path.
//不是并发安全的！ // Not concurrency-safe!
func (n *node) addRoute(path string, handlers HandlersChain) {
	faullPath := path
	n.priority++
	numParams := counParams(path)
	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(numParams, path, faullPath, handlers)
	}
}
