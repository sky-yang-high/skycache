package geecache

//兼容各种源数据的结构
type ByteView struct {
	b []byte
}

//实现 Value 接口
func (bv ByteView) Len() int {
	return len(bv.b)
}

//返回 源数据的 拷贝
func (bv ByteView) ByteSlice() []byte {
	return cloneByte(bv.b)
}

func (bv ByteView) String() string {
	return string(bv.b)
}

func cloneByte(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
