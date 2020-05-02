package gin_web

import (
	"io"
	"os"
)

var DefaultWriter io.Writer = os.Stdout

const(
	debugCode = iota
	releaseCode
	testCode
)

var ginMode = debugCode