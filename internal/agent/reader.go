package agent

import (
	"bufio"
	"context"
	"io"
)

type readerAgent struct {
	scanner  *bufio.Scanner
	exporter io.Writer
}

func NewReaderAgent(r io.Reader, exporter io.Writer) Agent {
	return &readerAgent{
		scanner:  bufio.NewScanner(r),
		exporter: exporter,
	}
}

// currently, we do not use ctx in this func because Scan run loop internaly until buffer found.
func (a *readerAgent) Run(ctx context.Context) error {
	for a.scanner.Scan() {
		if _, err := a.exporter.Write(a.scanner.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
