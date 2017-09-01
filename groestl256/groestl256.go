package groestl256

const (
	Size      = 32
	BlockSize = 4
)

type hash struct {
	nfactor int

	data []byte
	sum  []byte
}

func New() hash.Hash {
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

func Sum(data []byte) [Size]byte {
	panic("Not implemented.")
}
