package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type fileAgent struct {
	watcher     *fsnotify.Watcher
	fp          *os.File
	exporter    io.Writer
	exportSaved bool
}

func NewFileAgent(watcher *fsnotify.Watcher, fp *os.File, exporter io.Writer, exportSaved bool) (Agent, error) {
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

func (a *fileAgent) Run(ctx context.Context) error {
	b, err := io.ReadAll(a.fp)
	if err != nil {
		return err
	}
	if a.exportSaved {
		if _, err := a.exporter.Write(b); err != nil {
			return err
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-a.watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				b, err := io.ReadAll(a.fp)
				if err != nil {
					return err
				}
				if _, err := a.exporter.Write(b); err != nil {
					return err
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				return fmt.Errorf("file is removed or renamed or initialized. event: %s", event)
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
