package cache

type ByteView struct {
	data []byte
}

func (bv ByteView) Size() int {
	return len(bv.data)
}

func (bv ByteView) String() string {
	return string(bv.data)
}

func (bv ByteView) ByteSlice() []byte {
	return copyBytes(bv.data)
}

func copyBytes(b []byte) []byte {
	res := make([]byte, len(b))
	copy(res, b)
	return res
}
