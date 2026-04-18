package wiki

import "io"

type SanitizerReader struct {
	Reader io.Reader
}

var sanitizeByte = func() [256]byte {
	var t [256]byte
	for i := range t {
		b := byte(i)
		if (b < 0x20 && b != 0x09 && b != 0x0A && b != 0x0D) || b >= 0x80 {
			t[i] = ' '
		} else {
			t[i] = b
		}
	}
	return t
}()

func (s *SanitizerReader) Read(p []byte) (int, error) {
	n, err := s.Reader.Read(p)
	for i := 0; i < n; i++ {
		p[i] = sanitizeByte[p[i]]
	}
	return n, err
}
