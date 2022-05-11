package agent

import (
	"bufio"
	"os"

	"github.com/nakatamixi/cloud-logging-exporter/internal/exporter"
)

type stdinAgent struct {
	scanner  *bufio.Scanner
	exporter *exporter.Exporter
}

func NewStdinAgent(exporter *exporter.Exporter) Agent {
	return &stdinAgent{
		scanner:  bufio.NewScanner(os.Stdin),
		exporter: exporter,
	}
}
func (a *stdinAgent) Run() error {
	for a.scanner.Scan() {
		a.exporter.Export(a.scanner.Text())
	}
	return nil
}
