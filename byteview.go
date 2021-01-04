package ruoCache

//抽象了一个只读数据结构 ByteView 用来表示缓存值，主要的数据结构之一。
type ByteView struct {
	b []byte // 选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等。
}

// 返回view的长度
func (v ByteView) Len() int {
	return len(v.b)
}

// 
func (v *ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
