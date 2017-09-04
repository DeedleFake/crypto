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
			in: []byte{0x88, 0x80, 0x6e, 0x34, 0x84, 0x9e, 0xcc, 0x0e, 0xcf,
				0x22, 0x60, 0xce, 0xdf, 0xa1, 0x56, 0x37, 0x0e, 0x84,
				0x56, 0x17, 0xc6, 0x65, 0xea, 0xf7, 0x6d, 0x9f, 0xc9,
				0x44, 0x2b, 0x1f, 0xa3, 0x3d},
			out: "95677613350eb57f458ff267c5f602906ad4802f46ba9400431bb463773df5af",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := fmt.Sprintf("%x", Sum(test.in))
			if got != test.out {
				t.Errorf("Expected %q", test.out)
				t.Errorf("Got %q", got)
			}
		})
	}
}
