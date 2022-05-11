package exporter

import (
	"fmt"
	"log"

	"cloud.google.com/go/logging"
)

type Exporter struct {
	withStdout bool
	logger     *log.Logger
}

func NewExporter(client *logging.Client, withStdout bool, logName, severity string) (*Exporter, error) {
	var lseverity logging.Severity
	switch severity {
	case "default":
		lseverity = logging.Default
	case "debug":
		lseverity = logging.Debug
	case "info":
		lseverity = logging.Info
	case "notice":
		lseverity = logging.Notice
	case "warning":
		lseverity = logging.Warning
	case "error":
		lseverity = logging.Error
	case "critical":
		lseverity = logging.Critical
	case "alert":
		lseverity = logging.Alert
	case "emergency":
		lseverity = logging.Emergency
	default:
		return nil, fmt.Errorf("unknown severity: %s", severity)
	}
	logger := client.Logger(logName).StandardLogger(lseverity)
	return &Exporter{
		withStdout: withStdout,
		logger:     logger,
	}, nil
}

func (e *Exporter) Export(s string) {
	e.logger.Print(s)
	if e.withStdout {
		log.Print(s)
	}
}
