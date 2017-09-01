package groestl256

import (
	"hash"
)

const (
	Size      = 32
	BlockSize = 4
)

type impl struct {
	buf   [64]byte
	ptr   uintptr
	state [8]uint64
	count uint64

	data []byte
}

func New() hash.Hash {
	h := &impl{}
	h.Reset()
	return h
}

func (h *impl) Write(data []byte) (int, error) {
	h.data = append(h.data, data...)
	return len(data), nil
}

func (h *impl) Sum(data []byte) []byte {
	panic("Not implemented.")
}

func (h *impl) Reset() {
	*h = impl{}
	h.state[7] = Size * 8
}

func (h *impl) Size() int {
	return Size
}

func (h *impl) BlockSize() int {
	return BlockSize
}

func Sum(data []byte) (out [Size]byte) {
	h := New()
	h.Write(data)
	copy(out[:], h.Sum(nil))
	return
}
