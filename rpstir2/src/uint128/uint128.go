package uint128

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"strconv"
	"strings"
)

type Uint128 struct {
	High uint64
	Low  uint64
}

func NewUint128(i uint64) Uint128 {
	return Uint128{
		High: 0,
		Low:  i,
	}
}

func FromString(s string) Uint128 {
	u := NewUint128(1)
	u.SetString(s)
	return u
}

func FromBytes(b []byte) Uint128 {
	res := NewUint128(0)
	if len(b) < 17 {
		res.Low = binary.BigEndian.Uint64(b[0:])
	} else if len(b) < 33{
		res.Low = binary.BigEndian.Uint64(b[0:16])
		res.High = binary.BigEndian.Uint64(b[16:])
	} else {
		panic("overflow mfk")
	}
	return res
}

func (x Uint128) Lsh(i uint) Uint128 {
	switch true {
	case i > 127:
		return Uint128{
			High: 0,
			Low:  0,
		}
	case 128 > i && i > 63:
		x.High = x.Low & ((^uint64(0)) >> (i - 64)) << (i - 64)
		x.Low = 0
	case 64 > i && i > 0:
		tmp := ((^uint64(0)) >> (64 - i) << (64 - i)) & x.Low
		x.High = ((((^uint64(0)) >> i) & x.High) << i) | (tmp >> (64 - i))
		x.Low = (((^uint64(0)) >> i) & x.Low) << i
	}
	return x
}

func (x Uint128) Rsh(i uint) Uint128 {
	switch true {
	case i > 127:
		return Uint128{
			High: 0,
			Low:  0,
		}
	case 128 > i && i > 63:
		x.Low = x.High >> (i - 64)
		x.High = 0
	case 64 > i && i > 0:
		tmp := (x.High & ((^uint64(0)) >> (64 - i))) << (64 - i)
		x.High >>= i
		x.Low = ((x.Low) >> i) | tmp
	}
	return x
}

func (x Uint128) Add(y Uint128) Uint128 {
	res := Uint128{
		High: 0,
		Low:  0,
	}
	tmp := (x.Low & ((uint64(1)<<63)-1)) + (y.Low & ((uint64(1)<<63)-1))
	flag := (tmp >> 63) + (x.Low >> 63) + (y.Low >> 63)
	res.Low = (tmp & ((uint64(1)<<63)-1)) + ((flag & 1) << 63)
	res.High = x.High + y.High + (flag >> 1)
	return res
}

func (x Uint128) Sub(y Uint128) Uint128 {
	res := Uint128{
		High: 0,
		Low:  0,
	}

	if rtn := x.Cmp(y); rtn == 0 {
		return res
	} else if rtn == -1 {
		panic("surprise! uint128 sub overflows")
	}
	if x.Low > y.Low {
		res.High = x.High - y.High
		res.Low = x.Low - y.Low
	} else if x.Low < y.Low {
		res.Low = (^uint64(0)) - y.Low - x.Low + 1
		res.High = x.High - y.High - 1
	}
	return res
}

// Cmp compares x and y and returns:
// -1 if x < y
// +1 if x > y
func (x Uint128) Cmp(y Uint128) int {
	if highCmp := cmp(x.High, y.High); highCmp != 0 {
		return highCmp
	}
	if lowCmp := cmp(x.Low, y.Low); lowCmp != 0 {
		return lowCmp
	}
	return 0
}

func cmp(x, y uint64) int {
	switch {
	case x > y:
		return 1
	case x < y:
		return -1
	case x == y:
		return 0
	}
	return 0
}

func (x Uint128) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	if x.High > 0 {
		binary.Write(wr, binary.BigEndian, x.High)
	} else {
		binary.Write(wr, binary.BigEndian, uint64(0))
	}
	binary.Write(wr, binary.BigEndian, x.Low)
	return wr.Bytes()
}

func (x Uint128) String(base int) string {
	return big.NewInt(1).SetBytes(x.Bytes()).Text(base)
}

func (x Uint128) Uint64() uint64 {
	return x.Low
}

func (x *Uint128) SetString(s string) {
	if strings.ContainsAny(s, "xX") {
		s = s[2:]
	}
	if len(s) > 32 {
		panic("surprise, uint128 SetString overflow")
	} else if len(s) < 17 {
		x.High = 0
		x.Low, _ = strconv.ParseUint(s, 16,64)
	} else {
		x.High, _ = strconv.ParseUint(s[0:len(s)-16], 16, 64)
		x.Low, _ = strconv.ParseUint(s[len(s)-16:], 16, 64)
	}
}

func (x Uint128) Or(y Uint128) Uint128 {
	return Uint128{
		High: x.High | y.High,
		Low:  x.Low | y.Low,
	}
}

func (x Uint128) And(y Uint128) Uint128 {
	return Uint128{
		High: x.High & y.High,
		Low:  x.Low & y.Low,
	}
}

func (x Uint128) Xor(y Uint128) Uint128 {
	return Uint128{
		High: x.High ^ y.High,
		Low:  x.Low ^ y.Low,
	}
}

func (x Uint128) Not() Uint128 {
	return Uint128{
		High: ^x.High,
		Low:  ^x.Low,
	}
}