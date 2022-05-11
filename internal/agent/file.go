package agent

import (
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/nakatamixi/cloud-logging-exporter/internal/exporter"
)

type fileAgent struct {
	watcher     *fsnotify.Watcher
	fp          *os.File
	exporter    *exporter.Exporter
	exportSaved bool
}

func NewFileAgent(watcher *fsnotify.Watcher, fp *os.File, exporter *exporter.Exporter, exportSaved bool) (Agent, error) {
	if err := watcher.Add(fp.Name()); err != nil {
		return nil, err
	}
	return &fileAgent{
		watcher:     watcher,
		fp:          fp,
		exporter:    exporter,
		exportSaved: exportSaved,
	}, nil
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
