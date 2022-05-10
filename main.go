package main

import (
	"bufio"
	"context"
	"flag"
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
	logName     string
	file        string
	severity    string
	exportSaved bool
	touch       bool
)

func init() {
	flag.StringVar(&projectID, "project", "", "GCP project")
	flag.StringVar(&logName, "logname", "cloud-logging-exporter", "log name for Cloud Logging")
	flag.StringVar(&file, "file", "", "target file path for export")
	flag.StringVar(&severity, "severity", "info", "log severity(default|debug|info|notice|warning|error|critical|alert|emergency)")
	flag.BoolVar(&exportSaved, "saved", true, "export saved contents of file.")
	flag.BoolVar(&touch, "touch", false, "touch file if not exists.")
	flag.Parse()
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
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
		log.Fatalf("unknown severity: %s", severity)
	}
	logger := client.Logger(logName).StandardLogger(lseverity)

	if file == "" {
		stdin := bufio.NewScanner(os.Stdin)
		go func() {
			for stdin.Scan() {
				logger.Print(stdin.Text())
			}
		}()
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
		b, err := io.ReadAll(fp)
		if err != nil {
			log.Fatal(err)
		}
		if exportSaved && newFile {
			logger.Print(string(b))
			log.Print(string(b))
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						bt, err := io.ReadAll(fp)
						if err != nil {
							log.Fatal(err)
						}
						logger.Print(string(bt))
						log.Print(string(bt))
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()

		err = watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-ctx.Done()
	log.Println("completed")
}
