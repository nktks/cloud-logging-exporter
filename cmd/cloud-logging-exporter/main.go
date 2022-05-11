package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/logging"
	"github.com/fsnotify/fsnotify"
)

var (
	projectID string
	// logging filter is
	// logName="projects/$projectID/logs/$logName"
	logName       string
	file          string
	severity      string
	exportSaved   bool
	touch         bool
	withStdoutLog bool
)

func init() {
	flag.StringVar(&projectID, "project", "", "GCP project")
	flag.StringVar(&logName, "logname", "cloud-logging-exporter", "log name for Cloud Logging")
	flag.StringVar(&file, "file", "", "target file path for export")
	flag.StringVar(&severity, "severity", "info", "log severity(default|debug|info|notice|warning|error|critical|alert|emergency)")
	flag.BoolVar(&exportSaved, "saved", true, "export saved contents of file.")
	flag.BoolVar(&touch, "touch", false, "touch file if not exists.")
	flag.BoolVar(&withStdoutLog, "logstdout", false, "whether print log with stdout.")
	flag.Parse()
}

func main() {

	if projectID == "" {
		log.Fatal("need project flag.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	exporter, err := newExporter(client, withStdoutLog, logName, severity)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	var agent agent
	if file == "" {
		agent = newStdinAgent(exporter)
	} else {
		fp, err := os.Open(file)
		newFile := false
		if err != nil {
			if os.IsNotExist(err) && touch {
				fp, err = os.Create(file)
				if err != nil {
					log.Fatal(err)
				}
				newFile = true
			} else {
				log.Fatal(err)
			}
		}
		defer fp.Close()
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
		if err := watcher.Add(file); err != nil {
			log.Fatal(err)
		}

		agent = newFileAgent(watcher, fp, exporter, exportSaved && !newFile)
	}
	go func() {
		if err := agent.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	log.Println("completed")
}

type exporter struct {
	withStdout bool
	logger     *log.Logger
}

func newExporter(client *logging.Client, withStdout bool, logName, severity string) (*exporter, error) {
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
	return &exporter{
		withStdout: withStdout,
		logger:     logger,
	}, nil
}

func (e *exporter) Export(s string) {
	e.logger.Print(s)
	if e.withStdout {
		log.Print(s)
	}
}

type agent interface {
	Run() error
}

type fileAgent struct {
	watcher     *fsnotify.Watcher
	fp          *os.File
	exporter    *exporter
	exportSaved bool
}

func newFileAgent(watcher *fsnotify.Watcher, fp *os.File, exporter *exporter, exportSaved bool) agent {
	return &fileAgent{
		watcher:     watcher,
		fp:          fp,
		exporter:    exporter,
		exportSaved: exportSaved,
	}
}

func (a *fileAgent) Run() error {
	b, err := io.ReadAll(a.fp)
	if err != nil {
		return err
	}
	if a.exportSaved {
		a.exporter.Export(string(b))
	}
	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				b, err := io.ReadAll(a.fp)
				if err != nil {
					return err
				}
				a.exporter.Export(string(b))
			}
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return nil
			}
			log.Println("error:", err)
		}
	}
	return nil
}

type stdinAgent struct {
	scanner  *bufio.Scanner
	exporter *exporter
}

func newStdinAgent(exporter *exporter) agent {
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
