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
	ctx.state[7] = size
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
			var g, m [8]uint64
			for u := range ctx.state {
				m[u] = binary.BigEndian.Uint64(ctx.buf[u<<3:])
				g[u] = m[u] ^ ctx.state[u]
			}

			for r := 0; r < 10; r += 2 {
				roundSmallP(g[:], uint64(r))
				roundSmallP(g[:], uint64(r+1))
			}

			for r := 0; r < 10; r += 2 {
				roundSmallQ(m[:], uint64(r))
				roundSmallQ(m[:], uint64(r+1))
			}

			for u := range ctx.state {
				ctx.state[u] ^= g[u] ^ m[u]

				ctx.count++
				ctx.offset = 0
			}
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

	ctx.core(pad[:])

	x := ctx.state
	for r := 0; r < 10; r += 2 {
		roundSmallP(x[:], uint64(r))
		roundSmallP(x[:], uint64(r+1))
	}

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
	return (r << 56) ^ (^(j << 56))
}

func rstt(t []uint64, d uint64, a []uint64, b0, b1, b2, b3, b4, b5, b6, b7 uint64) {
	t[d] = t0[a[b0]>>56] ^ rotr64(t0[(a[b1]>>48)&0xFF], 8) ^ rotr64(t0[(a[b2]>>40)&0xFF], 16) ^ rotr64(t0[(a[b3]>>32)&0xFF], 24) ^ t4[(a[b4]>>24)&0xFF] ^ rotr64(t4[(a[b5]>>16)&0xFF], 8) ^ rotr64(t4[(a[b6]>>8)&0xFF], 16) ^ rotr64(t4[a[b7]&0xFF], 24)
}

func rotr64(x, n uint64) uint64 {
	return rotl64(x, 64-n)
}

func rotl64(x, n uint64) uint64 {
	return (x << n) | (x >> (64 - n))
}

var (
	t0 = []uint64{
		0xc632f4a5f497a5c6, 0xf86f978497eb84f8,
		0xee5eb099b0c799ee, 0xf67a8c8d8cf78df6,
		0xffe8170d17e50dff, 0xd60adcbddcb7bdd6,
		0xde16c8b1c8a7b1de, 0x916dfc54fc395491,
		0x6090f050f0c05060, 0x0207050305040302,
		0xce2ee0a9e087a9ce, 0x56d1877d87ac7d56,
		0xe7cc2b192bd519e7, 0xb513a662a67162b5,
		0x4d7c31e6319ae64d, 0xec59b59ab5c39aec,
		0x8f40cf45cf05458f, 0x1fa3bc9dbc3e9d1f,
		0x8949c040c0094089, 0xfa68928792ef87fa,
		0xefd03f153fc515ef, 0xb29426eb267febb2,
		0x8ece40c94007c98e, 0xfbe61d0b1ded0bfb,
		0x416e2fec2f82ec41, 0xb31aa967a97d67b3,
		0x5f431cfd1cbefd5f, 0x456025ea258aea45,
		0x23f9dabfda46bf23, 0x535102f702a6f753,
		0xe445a196a1d396e4, 0x9b76ed5bed2d5b9b,
		0x75285dc25deac275, 0xe1c5241c24d91ce1,
		0x3dd4e9aee97aae3d, 0x4cf2be6abe986a4c,
		0x6c82ee5aeed85a6c, 0x7ebdc341c3fc417e,
		0xf5f3060206f102f5, 0x8352d14fd11d4f83,
		0x688ce45ce4d05c68, 0x515607f407a2f451,
		0xd18d5c345cb934d1, 0xf9e1180818e908f9,
		0xe24cae93aedf93e2, 0xab3e9573954d73ab,
		0x6297f553f5c45362, 0x2a6b413f41543f2a,
		0x081c140c14100c08, 0x9563f652f6315295,
		0x46e9af65af8c6546, 0x9d7fe25ee2215e9d,
		0x3048782878602830, 0x37cff8a1f86ea137,
		0x0a1b110f11140f0a, 0x2febc4b5c45eb52f,
		0x0e151b091b1c090e, 0x247e5a365a483624,
		0x1badb69bb6369b1b, 0xdf98473d47a53ddf,
		0xcda76a266a8126cd, 0x4ef5bb69bb9c694e,
		0x7f334ccd4cfecd7f, 0xea50ba9fbacf9fea,
		0x123f2d1b2d241b12, 0x1da4b99eb93a9e1d,
		0x58c49c749cb07458, 0x3446722e72682e34,
		0x3641772d776c2d36, 0xdc11cdb2cda3b2dc,
		0xb49d29ee2973eeb4, 0x5b4d16fb16b6fb5b,
		0xa4a501f60153f6a4, 0x76a1d74dd7ec4d76,
		0xb714a361a37561b7, 0x7d3449ce49face7d,
		0x52df8d7b8da47b52, 0xdd9f423e42a13edd,
		0x5ecd937193bc715e, 0x13b1a297a2269713,
		0xa6a204f50457f5a6, 0xb901b868b86968b9,
		0x0000000000000000, 0xc1b5742c74992cc1,
		0x40e0a060a0806040, 0xe3c2211f21dd1fe3,
		0x793a43c843f2c879, 0xb69a2ced2c77edb6,
		0xd40dd9bed9b3bed4, 0x8d47ca46ca01468d,
		0x671770d970ced967, 0x72afdd4bdde44b72,
		0x94ed79de7933de94, 0x98ff67d4672bd498,
		0xb09323e8237be8b0, 0x855bde4ade114a85,
		0xbb06bd6bbd6d6bbb, 0xc5bb7e2a7e912ac5,
		0x4f7b34e5349ee54f, 0xedd73a163ac116ed,
		0x86d254c55417c586, 0x9af862d7622fd79a,
		0x6699ff55ffcc5566, 0x11b6a794a7229411,
		0x8ac04acf4a0fcf8a, 0xe9d9301030c910e9,
		0x040e0a060a080604, 0xfe66988198e781fe,
		0xa0ab0bf00b5bf0a0, 0x78b4cc44ccf04478,
		0x25f0d5bad54aba25, 0x4b753ee33e96e34b,
		0xa2ac0ef30e5ff3a2, 0x5d4419fe19bafe5d,
		0x80db5bc05b1bc080, 0x0580858a850a8a05,
		0x3fd3ecadec7ead3f, 0x21fedfbcdf42bc21,
		0x70a8d848d8e04870, 0xf1fd0c040cf904f1,
		0x63197adf7ac6df63, 0x772f58c158eec177,
		0xaf309f759f4575af, 0x42e7a563a5846342,
		0x2070503050403020, 0xe5cb2e1a2ed11ae5,
		0xfdef120e12e10efd, 0xbf08b76db7656dbf,
		0x8155d44cd4194c81, 0x18243c143c301418,
		0x26795f355f4c3526, 0xc3b2712f719d2fc3,
		0xbe8638e13867e1be, 0x35c8fda2fd6aa235,
		0x88c74fcc4f0bcc88, 0x2e654b394b5c392e,
		0x936af957f93d5793, 0x55580df20daaf255,
		0xfc619d829de382fc, 0x7ab3c947c9f4477a,
		0xc827efacef8bacc8, 0xba8832e7326fe7ba,
		0x324f7d2b7d642b32, 0xe642a495a4d795e6,
		0xc03bfba0fb9ba0c0, 0x19aab398b3329819,
		0x9ef668d16827d19e, 0xa322817f815d7fa3,
		0x44eeaa66aa886644, 0x54d6827e82a87e54,
		0x3bdde6abe676ab3b, 0x0b959e839e16830b,
		0x8cc945ca4503ca8c, 0xc7bc7b297b9529c7,
		0x6b056ed36ed6d36b, 0x286c443c44503c28,
		0xa72c8b798b5579a7, 0xbc813de23d63e2bc,
		0x1631271d272c1d16, 0xad379a769a4176ad,
		0xdb964d3b4dad3bdb, 0x649efa56fac85664,
		0x74a6d24ed2e84e74, 0x1436221e22281e14,
		0x92e476db763fdb92, 0x0c121e0a1e180a0c,
		0x48fcb46cb4906c48, 0xb88f37e4376be4b8,
		0x9f78e75de7255d9f, 0xbd0fb26eb2616ebd,
		0x43692aef2a86ef43, 0xc435f1a6f193a6c4,
		0x39dae3a8e372a839, 0x31c6f7a4f762a431,
		0xd38a593759bd37d3, 0xf274868b86ff8bf2,
		0xd583563256b132d5, 0x8b4ec543c50d438b,
		0x6e85eb59ebdc596e, 0xda18c2b7c2afb7da,
		0x018e8f8c8f028c01, 0xb11dac64ac7964b1,
		0x9cf16dd26d23d29c, 0x49723be03b92e049,
		0xd81fc7b4c7abb4d8, 0xacb915fa1543faac,
		0xf3fa090709fd07f3, 0xcfa06f256f8525cf,
		0xca20eaafea8fafca, 0xf47d898e89f38ef4,
		0x476720e9208ee947, 0x1038281828201810,
		0x6f0b64d564ded56f, 0xf073838883fb88f0,
		0x4afbb16fb1946f4a, 0x5cca967296b8725c,
		0x38546c246c702438, 0x575f08f108aef157,
		0x732152c752e6c773, 0x9764f351f3355197,
		0xcbae6523658d23cb, 0xa125847c84597ca1,
		0xe857bf9cbfcb9ce8, 0x3e5d6321637c213e,
		0x96ea7cdd7c37dd96, 0x611e7fdc7fc2dc61,
		0x0d9c9186911a860d, 0x0f9b9485941e850f,
		0xe04bab90abdb90e0, 0x7cbac642c6f8427c,
		0x712657c457e2c471, 0xcc29e5aae583aacc,
		0x90e373d8733bd890, 0x06090f050f0c0506,
		0xf7f4030103f501f7, 0x1c2a36123638121c,
		0xc23cfea3fe9fa3c2, 0x6a8be15fe1d45f6a,
		0xaebe10f91047f9ae, 0x69026bd06bd2d069,
		0x17bfa891a82e9117, 0x9971e858e8295899,
		0x3a5369276974273a, 0x27f7d0b9d04eb927,
		0xd991483848a938d9, 0xebde351335cd13eb,
		0x2be5ceb3ce56b32b, 0x2277553355443322,
		0xd204d6bbd6bfbbd2, 0xa9399070904970a9,
		0x07878089800e8907, 0x33c1f2a7f266a733,
		0x2decc1b6c15ab62d, 0x3c5a66226678223c,
		0x15b8ad92ad2a9215, 0xc9a96020608920c9,
		0x875cdb49db154987, 0xaab01aff1a4fffaa,
		0x50d8887888a07850, 0xa52b8e7a8e517aa5,
		0x03898a8f8a068f03, 0x594a13f813b2f859,
		0x09929b809b128009, 0x1a2339173934171a,
		0x651075da75cada65, 0xd784533153b531d7,
		0x84d551c65113c684, 0xd003d3b8d3bbb8d0,
		0x82dc5ec35e1fc382, 0x29e2cbb0cb52b029,
		0x5ac3997799b4775a, 0x1e2d3311333c111e,
		0x7b3d46cb46f6cb7b, 0xa8b71ffc1f4bfca8,
		0x6d0c61d661dad66d, 0x2c624e3a4e583a2c,
	}

	t4 = []uint64{
		0xf497a5c6c632f4a5, 0x97eb84f8f86f9784,
		0xb0c799eeee5eb099, 0x8cf78df6f67a8c8d,
		0x17e50dffffe8170d, 0xdcb7bdd6d60adcbd,
		0xc8a7b1dede16c8b1, 0xfc395491916dfc54,
		0xf0c050606090f050, 0x0504030202070503,
		0xe087a9cece2ee0a9, 0x87ac7d5656d1877d,
		0x2bd519e7e7cc2b19, 0xa67162b5b513a662,
		0x319ae64d4d7c31e6, 0xb5c39aecec59b59a,
		0xcf05458f8f40cf45, 0xbc3e9d1f1fa3bc9d,
		0xc00940898949c040, 0x92ef87fafa689287,
		0x3fc515efefd03f15, 0x267febb2b29426eb,
		0x4007c98e8ece40c9, 0x1ded0bfbfbe61d0b,
		0x2f82ec41416e2fec, 0xa97d67b3b31aa967,
		0x1cbefd5f5f431cfd, 0x258aea45456025ea,
		0xda46bf2323f9dabf, 0x02a6f753535102f7,
		0xa1d396e4e445a196, 0xed2d5b9b9b76ed5b,
		0x5deac27575285dc2, 0x24d91ce1e1c5241c,
		0xe97aae3d3dd4e9ae, 0xbe986a4c4cf2be6a,
		0xeed85a6c6c82ee5a, 0xc3fc417e7ebdc341,
		0x06f102f5f5f30602, 0xd11d4f838352d14f,
		0xe4d05c68688ce45c, 0x07a2f451515607f4,
		0x5cb934d1d18d5c34, 0x18e908f9f9e11808,
		0xaedf93e2e24cae93, 0x954d73abab3e9573,
		0xf5c453626297f553, 0x41543f2a2a6b413f,
		0x14100c08081c140c, 0xf63152959563f652,
		0xaf8c654646e9af65, 0xe2215e9d9d7fe25e,
		0x7860283030487828, 0xf86ea13737cff8a1,
		0x11140f0a0a1b110f, 0xc45eb52f2febc4b5,
		0x1b1c090e0e151b09, 0x5a483624247e5a36,
		0xb6369b1b1badb69b, 0x47a53ddfdf98473d,
		0x6a8126cdcda76a26, 0xbb9c694e4ef5bb69,
		0x4cfecd7f7f334ccd, 0xbacf9feaea50ba9f,
		0x2d241b12123f2d1b, 0xb93a9e1d1da4b99e,
		0x9cb0745858c49c74, 0x72682e343446722e,
		0x776c2d363641772d, 0xcda3b2dcdc11cdb2,
		0x2973eeb4b49d29ee, 0x16b6fb5b5b4d16fb,
		0x0153f6a4a4a501f6, 0xd7ec4d7676a1d74d,
		0xa37561b7b714a361, 0x49face7d7d3449ce,
		0x8da47b5252df8d7b, 0x42a13edddd9f423e,
		0x93bc715e5ecd9371, 0xa226971313b1a297,
		0x0457f5a6a6a204f5, 0xb86968b9b901b868,
		0x0000000000000000, 0x74992cc1c1b5742c,
		0xa080604040e0a060, 0x21dd1fe3e3c2211f,
		0x43f2c879793a43c8, 0x2c77edb6b69a2ced,
		0xd9b3bed4d40dd9be, 0xca01468d8d47ca46,
		0x70ced967671770d9, 0xdde44b7272afdd4b,
		0x7933de9494ed79de, 0x672bd49898ff67d4,
		0x237be8b0b09323e8, 0xde114a85855bde4a,
		0xbd6d6bbbbb06bd6b, 0x7e912ac5c5bb7e2a,
		0x349ee54f4f7b34e5, 0x3ac116ededd73a16,
		0x5417c58686d254c5, 0x622fd79a9af862d7,
		0xffcc55666699ff55, 0xa722941111b6a794,
		0x4a0fcf8a8ac04acf, 0x30c910e9e9d93010,
		0x0a080604040e0a06, 0x98e781fefe669881,
		0x0b5bf0a0a0ab0bf0, 0xccf0447878b4cc44,
		0xd54aba2525f0d5ba, 0x3e96e34b4b753ee3,
		0x0e5ff3a2a2ac0ef3, 0x19bafe5d5d4419fe,
		0x5b1bc08080db5bc0, 0x850a8a050580858a,
		0xec7ead3f3fd3ecad, 0xdf42bc2121fedfbc,
		0xd8e0487070a8d848, 0x0cf904f1f1fd0c04,
		0x7ac6df6363197adf, 0x58eec177772f58c1,
		0x9f4575afaf309f75, 0xa584634242e7a563,
		0x5040302020705030, 0x2ed11ae5e5cb2e1a,
		0x12e10efdfdef120e, 0xb7656dbfbf08b76d,
		0xd4194c818155d44c, 0x3c30141818243c14,
		0x5f4c352626795f35, 0x719d2fc3c3b2712f,
		0x3867e1bebe8638e1, 0xfd6aa23535c8fda2,
		0x4f0bcc8888c74fcc, 0x4b5c392e2e654b39,
		0xf93d5793936af957, 0x0daaf25555580df2,
		0x9de382fcfc619d82, 0xc9f4477a7ab3c947,
		0xef8bacc8c827efac, 0x326fe7baba8832e7,
		0x7d642b32324f7d2b, 0xa4d795e6e642a495,
		0xfb9ba0c0c03bfba0, 0xb332981919aab398,
		0x6827d19e9ef668d1, 0x815d7fa3a322817f,
		0xaa88664444eeaa66, 0x82a87e5454d6827e,
		0xe676ab3b3bdde6ab, 0x9e16830b0b959e83,
		0x4503ca8c8cc945ca, 0x7b9529c7c7bc7b29,
		0x6ed6d36b6b056ed3, 0x44503c28286c443c,
		0x8b5579a7a72c8b79, 0x3d63e2bcbc813de2,
		0x272c1d161631271d, 0x9a4176adad379a76,
		0x4dad3bdbdb964d3b, 0xfac85664649efa56,
		0xd2e84e7474a6d24e, 0x22281e141436221e,
		0x763fdb9292e476db, 0x1e180a0c0c121e0a,
		0xb4906c4848fcb46c, 0x376be4b8b88f37e4,
		0xe7255d9f9f78e75d, 0xb2616ebdbd0fb26e,
		0x2a86ef4343692aef, 0xf193a6c4c435f1a6,
		0xe372a83939dae3a8, 0xf762a43131c6f7a4,
		0x59bd37d3d38a5937, 0x86ff8bf2f274868b,
		0x56b132d5d5835632, 0xc50d438b8b4ec543,
		0xebdc596e6e85eb59, 0xc2afb7dada18c2b7,
		0x8f028c01018e8f8c, 0xac7964b1b11dac64,
		0x6d23d29c9cf16dd2, 0x3b92e04949723be0,
		0xc7abb4d8d81fc7b4, 0x1543faacacb915fa,
		0x09fd07f3f3fa0907, 0x6f8525cfcfa06f25,
		0xea8fafcaca20eaaf, 0x89f38ef4f47d898e,
		0x208ee947476720e9, 0x2820181010382818,
		0x64ded56f6f0b64d5, 0x83fb88f0f0738388,
		0xb1946f4a4afbb16f, 0x96b8725c5cca9672,
		0x6c70243838546c24, 0x08aef157575f08f1,
		0x52e6c773732152c7, 0xf33551979764f351,
		0x658d23cbcbae6523, 0x84597ca1a125847c,
		0xbfcb9ce8e857bf9c, 0x637c213e3e5d6321,
		0x7c37dd9696ea7cdd, 0x7fc2dc61611e7fdc,
		0x911a860d0d9c9186, 0x941e850f0f9b9485,
		0xabdb90e0e04bab90, 0xc6f8427c7cbac642,
		0x57e2c471712657c4, 0xe583aacccc29e5aa,
		0x733bd89090e373d8, 0x0f0c050606090f05,
		0x03f501f7f7f40301, 0x3638121c1c2a3612,
		0xfe9fa3c2c23cfea3, 0xe1d45f6a6a8be15f,
		0x1047f9aeaebe10f9, 0x6bd2d06969026bd0,
		0xa82e911717bfa891, 0xe82958999971e858,
		0x6974273a3a536927, 0xd04eb92727f7d0b9,
		0x48a938d9d9914838, 0x35cd13ebebde3513,
		0xce56b32b2be5ceb3, 0x5544332222775533,
		0xd6bfbbd2d204d6bb, 0x904970a9a9399070,
		0x800e890707878089, 0xf266a73333c1f2a7,
		0xc15ab62d2decc1b6, 0x6678223c3c5a6622,
		0xad2a921515b8ad92, 0x608920c9c9a96020,
		0xdb154987875cdb49, 0x1a4fffaaaab01aff,
		0x88a0785050d88878, 0x8e517aa5a52b8e7a,
		0x8a068f0303898a8f, 0x13b2f859594a13f8,
		0x9b12800909929b80, 0x3934171a1a233917,
		0x75cada65651075da, 0x53b531d7d7845331,
		0x5113c68484d551c6, 0xd3bbb8d0d003d3b8,
		0x5e1fc38282dc5ec3, 0xcb52b02929e2cbb0,
		0x99b4775a5ac39977, 0x333c111e1e2d3311,
		0x46f6cb7b7b3d46cb, 0x1f4bfca8a8b71ffc,
		0x61dad66d6d0c61d6, 0x4e583a2c2c624e3a,
	}
)
