package randomreader

import (
	"io"
	"math/rand"
)

type RandomTextReader struct {
	remainingChars int64
}

func New(length int64) *RandomTextReader {
	return &RandomTextReader{
		remainingChars: length,
	}
}

func (r *RandomTextReader) Read(p []byte) (n int, err error) {
	const chars = "abcdefghijklmnopqrstuvwxyz "
	if r.remainingChars == 0 {
		return 0, io.EOF
	}
	writeChars := func() int {
		if r.remainingChars < int64(len(p)) {
			return int(r.remainingChars)
		} else {
			return len(p)
		}
	}()
	r.remainingChars = r.remainingChars - int64(writeChars)
	for i := 0; i < writeChars; i++ {
		p[i] = chars[rand.Intn(len(chars))]
	}
	return writeChars, nil
}
