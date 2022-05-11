package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/logging"
	"github.com/fsnotify/fsnotify"
	"github.com/nakatamixi/cloud-logging-exporter/internal/agent"
	"github.com/nakatamixi/cloud-logging-exporter/internal/exporter"
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
	exporter, err := exporter.NewExporter(client, withStdoutLog, logName, severity)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	var a agent.Agent
	if file == "" {
		a = agent.NewStdinAgent(exporter)
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

		a = agent.NewFileAgent(watcher, fp, exporter, exportSaved && !newFile)
	}
	go func() {
		if err := a.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	log.Println("completed")
}
