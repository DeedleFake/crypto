package groestl256

import (
	"fmt"
	"testing"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		out  string
	}{
		{
			name: "Basic",
			in:   []byte{0x34, 0x6e, 0x80, 0x88, 0x0e, 0xcc, 0x9e, 0x84, 0xce, 0x60, 0x22, 0xcf, 0x37, 0x56, 0xa1, 0xdf, 0x17, 0x56, 0x84, 0x0e, 0xf7, 0xea, 0x65, 0xc6, 0x44, 0xc9, 0x9f, 0x6d, 0x3d, 0xa3, 0x1f, 0x2b},
			out:  "137667957fb50e3567f28f459002f6c52f80d46a0094ba4663b41b43aff53d77",
		},
		{
			name: "Basic2",
			in:   []byte{0x34, 0x6e, 0x80, 0x88, 0x0e, 0xcc, 0x9e, 0x84, 0xce, 0x60, 0x22, 0xcf, 0x37, 0x56, 0xa1, 0xdf, 0x17, 0x56, 0x84, 0x0e, 0xf7, 0xea, 0x65, 0xc6, 0x44, 0xc9, 0x9f, 0x6d, 0x3d, 0xa3, 0x1f, 0x2b},
			out:  "137667957fb50e3567f28f459002f6c52f80d46a0094ba4663b41b43aff53d77",
		},
	}

	h := New()
	var out [Size]byte

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			h.Write(test.in)
			h.Sum(out[:0])

			got := fmt.Sprintf("%x", out)
			if got != test.out {
				t.Errorf("Expected %q", test.out)
				t.Errorf("Got %q", got)
			}
		})
	}
}
