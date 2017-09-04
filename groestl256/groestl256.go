package groestl256

import (
	"encoding/binary"
	"hash"
)

const (
	Size      = 32
	BlockSize = 64
)

type context struct {
	buf    [64]byte
	offset int
	state  [8]uint64
	count  uint64
}

func newContext(size uint64) *context {
	ctx := &context{}
	ctx.state[7] = size * 8
	return ctx
}

func (ctx *context) core(data []byte) {
	if len(data) < len(ctx.buf)-ctx.offset {
		copy(ctx.buf[ctx.offset:], data)
		ctx.offset += len(data)
		return
	}

	for len(data) > 0 {
		clen := len(ctx.buf) - ctx.offset
		if clen > len(data) {
			clen = len(data)
		}

		copy(ctx.buf[ctx.offset:], data[:clen])
		ctx.offset += clen
		data = data[clen:]

		if ctx.offset == len(ctx.buf) {
			var g, m [len(ctx.state)]uint64
			for u := range ctx.state {
				m[u] = binary.BigEndian.Uint64(ctx.buf[u<<3:])
				g[u] = m[u] ^ ctx.state[u]
			}

			permSmallP(g[:])
			permSmallQ(m[:])

			for u := range ctx.state {
				ctx.state[u] ^= g[u] ^ m[u]
			}

			ctx.count++
			ctx.offset = 0
		}
	}
}

func (ctx *context) close(dst []byte, ub, n uint64) {
	var pad [72]byte

	z := uint64(0x80) >> n
	pad[0] = uint8((ub & -z) | z)

	padLen := 128 - ctx.offset
	count := ctx.count + 2
	if ctx.offset < 56 {
		padLen = 64 - ctx.offset
		count = ctx.count + 1
	}
	binary.BigEndian.PutUint64(pad[padLen-8:], count)

	ctx.core(pad[:padLen])

	x := ctx.state
	permSmallP(x[:])

	for u := range x {
		ctx.state[u] ^= x[u]
	}
	for u := 0; u < 4; u++ {
		binary.BigEndian.PutUint64(pad[u<<3:], ctx.state[u+4])
	}

	copy(dst, pad[32-len(dst):])

	//ctx.init(uint64(len(dst)) << 3)
}

type impl struct {
	data []byte
}

func New() hash.Hash {
	return &impl{}
}

func (h *impl) Write(data []byte) (int, error) {
	h.data = append(h.data, data...)
	return len(data), nil
}

func (h *impl) Sum(prev []byte) []byte {
	out := append(prev, make([]byte, Size)...)

	ctx := newContext(Size)
	ctx.core(h.data)
	ctx.close(out[len(prev):], 0, 0)

	return out
}

func (h *impl) Reset() {
	h.data = h.data[:0]
}

func (h *impl) Size() int {
	return Size
}

func (h *impl) BlockSize() int {
	return BlockSize
}

func Sum(data []byte) (out [Size]byte) {
	ctx := newContext(Size)
	ctx.core(data)
	ctx.close(out[:], 0, 0)
	return
}
