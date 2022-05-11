package agent

import (
	"bufio"
	"io"

	"github.com/nakatamixi/cloud-logging-exporter/internal/exporter"
)

type readerAgent struct {
	scanner  *bufio.Scanner
	exporter *exporter.Exporter
}

func NewReaderAgent(r io.Reader, exporter *exporter.Exporter) Agent {
	return &readerAgent{
		scanner:  bufio.NewScanner(r),
		exporter: exporter,
	}
}
func (a *readerAgent) Run() error {
	for a.scanner.Scan() {
		a.exporter.Export(a.scanner.Text())
	}
	return nil
}
