package ttf

import (
	"encoding/binary"
	"io"
	"os"
)

type Fixed struct {
	base int16
	frac uint16
}

func (f *Fixed) Read(file io.Reader) os.Error {
	return readValues(file, &f.base, &f.frac)
}

func (f *Fixed) Tof64() float64 {
	return float64(f.base) + float64(f.frac)/65536
}

type FWord int16
type uFWord uint16

type longDateTime int64

type PANOSE [10]byte

func readValues(r io.Reader, values ...interface{}) (err os.Error) {
	for _, v := range values {
		if err = binary.Read(r, binary.BigEndian, v); err != nil {
			return
		}
	}
	return
}
