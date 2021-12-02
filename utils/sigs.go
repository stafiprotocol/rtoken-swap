package utils

func DecomposeSignature(sigs [][]byte) ([]uint8, [][32]byte, [][32]byte) {
	rs := make([][32]byte, 0)
	ss := make([][32]byte, 0)
	vs := make([]uint8, 0)

	for _, sig := range sigs {
		var r [32]byte
		var s [32]byte
		copy(r[:], sig[:32])
		copy(s[:], sig[32:64])
		rs = append(rs, r)
		ss = append(ss, s)
		vs = append(vs, sig[64:][0])
	}

	return vs, rs, ss
}
