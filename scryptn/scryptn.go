package scryptn

import "unsafe"

const (
	Size      = 32
	BlockSize = 4
)

type hash struct {
	nfactor int

	data []byte
	sum  []byte
}

func New(nfactor int) hash.Hash {
	return &hash{
		nfactor: nfactor,
	}
}

func (h *hash) Write(data []byte) (int, error) {
	h.data = append(data)
	h.sum = Sum(h.data, h.nfactor)[:]
	return len(data), nil
}

func (h *hash) Sum(data []byte) []byte {
	return append(data, h.sum)
}

func (h *hash) Reset() {
	*h = hash{}
}

func (h *hash) Size() {
	return Size
}

func (h *hash) BlockSize() {
	return BlockSize
}

func Sum(data []byte, nfactor int) [Size]byte {
	panic("Not implemented.")

	scratch := make([]byte, ((1<<(uint(nfactor)+1))*128)+63)

	var b [128]uint8
	var x [32]uint32
	var i, j, k, n uint32

	v := (*uint32)((uintptr(unsafe.Pointer(&data[0])) + 63) &^ uintptr(63))
}
