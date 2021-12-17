package ctx

import (
	"io"
	"os"
)

type ProxyWriter struct {
	writer io.Writer
	spy    func(chunk *[]byte) error
}

func (proxy ProxyWriter) Write(chunk []byte) (n int, err error) {
	err = proxy.spy(&chunk)
	if err != nil {
		return 0, err
	}
	return proxy.writer.Write(chunk)
}

func NewSystemOutProxyWriter(writer io.Writer) io.Writer {
	return ProxyWriter{
		writer: writer,
		spy: func(chunk *[]byte) error {
			_, err := os.Stdout.Write(*chunk)
			return err
		},
	}
}
