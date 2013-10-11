package daemon

import (
	"io"
)

const MAX_PIPE_BYTES = 32

type Pipe struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func NewPipe(reader io.ReadCloser, writer io.WriteCloser) *Pipe {
	return &Pipe{
		reader: reader,
		writer: writer,
	}
}

func (p *Pipe) Read(b []byte) (n int, err error) {
	return p.reader.Read(b)
}

func (p *Pipe) Write(b []byte) (n int, err error) {
	return p.writer.Write(b)
}

func (p *Pipe) Close() error {
	p.reader.Close()
	p.writer.Close()

	return nil
}
