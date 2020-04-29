package gin_web

// Param 参数是单个URL参数，由键和值组成
type Param struct {
	Key string
	Value string
}

// 参数是路由器返回的参数切片
// 切片是有序的，第一个URL参数也是第一个切片值
// 因此，通过索引读取值是安全的
type Params []Param
