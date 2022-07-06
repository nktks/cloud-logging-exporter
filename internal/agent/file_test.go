package agent_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"

	"github.com/nktks/cloud-logging-exporter/internal/agent"
)

func TestFileAgent_Run(t *testing.T) {
	t.Run("export saved", func(t *testing.T) {
		saved := "saved"
		file, err := os.CreateTemp("./", "")
		assert.NoError(t, err)
		defer os.Remove(file.Name())
		_, err = file.Write([]byte(saved))
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		watcher, err := fsnotify.NewWatcher()
		assert.NoError(t, err)
		defer watcher.Close()
		mocke := strings.Builder{}
		fp, err := os.Open(file.Name())
		assert.NoError(t, err)
		a, err := agent.NewFileAgent(watcher, fp, &mocke, true)
		assert.NoError(t, err)
		go func() {
			assert.NoError(t, a.Run(ctx))
		}()
		time.Sleep(1 * time.Second)
		cancel()
		assert.Equal(t, saved, mocke.String())
	})
	t.Run("not export saved", func(t *testing.T) {
		saved := "saved"
		file, err := os.CreateTemp("./", "")
		assert.NoError(t, err)
		defer os.Remove(file.Name())
		_, err = file.Write([]byte(saved))
		assert.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		watcher, err := fsnotify.NewWatcher()
		assert.NoError(t, err)
		defer watcher.Close()
		mocke := strings.Builder{}
		fp, err := os.Open(file.Name())
		assert.NoError(t, err)
		a, err := agent.NewFileAgent(watcher, fp, &mocke, false)
		assert.NoError(t, err)
		go func() {
			assert.NoError(t, a.Run(ctx))
		}()
		time.Sleep(1 * time.Second)

		appended := "appended"

		_, err = file.Write([]byte(appended))
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)
		cancel()
		assert.Equal(t, appended, mocke.String())
	})
	t.Run("file removed", func(t *testing.T) {
		saved := "saved"
		file, err := os.CreateTemp("./", "")
		assert.NoError(t, err)
		_, err = file.Write([]byte(saved))
		assert.NoError(t, err)

		ctx, _ := context.WithCancel(context.Background())

		watcher, err := fsnotify.NewWatcher()
		assert.NoError(t, err)
		defer watcher.Close()
		mocke := strings.Builder{}
		fp, err := os.Open(file.Name())
		assert.NoError(t, err)
		a, err := agent.NewFileAgent(watcher, fp, &mocke, false)
		assert.NoError(t, err)
		os.Remove(file.Name())
		assert.Error(t, a.Run(ctx))
	})
}
