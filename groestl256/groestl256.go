package groestl256

import (
	"encoding/binary"
	"hash"
)

const (
	Size      = 32
	BlockSize = 64
)

type impl struct {
	buf    [64]byte
	offset int
	state  [8]uint64
	count  uint64

	data []byte
}

func New() hash.Hash {
	h := &impl{}
	h.Reset()
	return h
}

func (h *impl) stateBytes() (out []byte) {
	panic("Not implemented.")
}

func (h *impl) Write(data []byte) (int, error) {
	h.data = append(h.data, data...)
	return len(data), nil
}

func (h *impl) Sum(prev []byte) []byte {
	data := h.data

	if len(data) < len(h.buf)-h.offset {
		copy(h.buf[h.offset:], data)
		h.offset += len(data)
	}

	for len(data) > 0 {
		clen := len(h.buf) - h.offset
		if clen > len(data) {
			clen = len(data)
		}

		copy(h.buf[h.offset:], data[:clen])
		h.offset += clen
		data = data[clen:]

		if h.offset == len(h.buf) {
			var g, m [8]uint64
			for u := 0; u < 8; u++ {
				m[u] = binary.BigEndian.Uint64(h.buf[u<<3:])
				g[u] = m[u] ^ h.state[u]
			}

			for r := 0; r < 10; r += 2 {
				roundSmallP(g[:], uint64(r))
				roundSmallP(g[:], uint64(r+1))
			}

			for r := 0; r < 10; r += 2 {
				roundSmallQ(m[:], uint64(r))
				roundSmallQ(m[:], uint64(r+1))
			}

			for u := 0; u < 8; u++ {
				h.state[u] ^= g[u] ^ m[u]
			}

			h.count++
			h.offset = 0
		}
	}

	return append(prev, h.stateBytes()...)
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

func roundSmallP(a []uint64, r uint64) {
	var t [8]uint64
	a[0] ^= pc64(0x00, r)
	a[1] ^= pc64(0x10, r)
	a[2] ^= pc64(0x20, r)
	a[3] ^= pc64(0x30, r)
	a[4] ^= pc64(0x40, r)
	a[5] ^= pc64(0x50, r)
	a[6] ^= pc64(0x60, r)
	a[7] ^= pc64(0x70, r)
	rstt(t[:], 0, a, 0, 1, 2, 3, 4, 5, 6, 7)
	rstt(t[:], 1, a, 1, 2, 3, 4, 5, 6, 7, 0)
	rstt(t[:], 2, a, 2, 3, 4, 5, 6, 7, 0, 1)
	rstt(t[:], 3, a, 3, 4, 5, 6, 7, 0, 1, 2)
	rstt(t[:], 4, a, 4, 5, 6, 7, 0, 1, 2, 3)
	rstt(t[:], 5, a, 5, 6, 7, 0, 1, 2, 3, 4)
	rstt(t[:], 6, a, 6, 7, 0, 1, 2, 3, 4, 5)
	rstt(t[:], 7, a, 7, 0, 1, 2, 3, 4, 5, 6)
	a[0] = t[0]
	a[1] = t[1]
	a[2] = t[2]
	a[3] = t[3]
	a[4] = t[4]
	a[5] = t[5]
	a[6] = t[6]
	a[7] = t[7]
}

func roundSmallQ(a []uint64, r uint64) {
	var t [8]uint64
	a[0] ^= qc64(0x00, r)
	a[1] ^= qc64(0x10, r)
	a[2] ^= qc64(0x20, r)
	a[3] ^= qc64(0x30, r)
	a[4] ^= qc64(0x40, r)
	a[5] ^= qc64(0x50, r)
	a[6] ^= qc64(0x60, r)
	a[7] ^= qc64(0x70, r)
	rstt(t[:], 0, a, 1, 3, 5, 7, 0, 2, 4, 6)
	rstt(t[:], 1, a, 2, 4, 6, 0, 1, 3, 5, 7)
	rstt(t[:], 2, a, 3, 5, 7, 1, 2, 4, 6, 0)
	rstt(t[:], 3, a, 4, 6, 0, 2, 3, 5, 7, 1)
	rstt(t[:], 4, a, 5, 7, 1, 3, 4, 6, 0, 2)
	rstt(t[:], 5, a, 6, 0, 2, 4, 5, 7, 1, 3)
	rstt(t[:], 6, a, 7, 1, 3, 5, 6, 0, 2, 4)
	rstt(t[:], 7, a, 0, 2, 4, 6, 7, 1, 3, 5)
	a[0] = t[0]
	a[1] = t[1]
	a[2] = t[2]
	a[3] = t[3]
	a[4] = t[4]
	a[5] = t[5]
	a[6] = t[6]
	a[7] = t[7]
}

func pc64(j, r uint64) uint64 {
	return j + r
}

func qc64(j, r uint64) uint64 {
	return (r << 56) ^ sphT64(^(j << 56))
}

func rstt(t []uint64, d uint64, a []uint64, b0, b01, b2, b3, b4, b5, b6, b7 uint64) {
	t[d] = T0[B64_0(a[b0])] ^ R64(T0[B64_1(a[b1])], 8) ^ R64(T0[B64_2(a[b2])], 16) ^ R64(T0[B64_3(a[b3])], 24) ^ T4[B64_4(a[b4])] ^ R64(T4[B64_5(a[b5])], 8) ^ R64(T4[B64_6(a[b6])], 16) ^ R64(T4[B64_7(a[b7])], 24)
}
