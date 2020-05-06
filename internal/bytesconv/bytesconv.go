package bytesconv

import "unsafe"

// BytesToString将字节片转换为字符串，而不分配内存。 // BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string()(unsafe.Pointer(&b)))
}
