package log

import (
	"io"
)

type MultiWriter struct {
	writers []io.Writer
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)

		if err != nil {
			return
		}
	}

	return
}

func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

func (mw *MultiWriter) AddWriter(w io.Writer) {
	mw.writers = append(mw.writers, w)
}
