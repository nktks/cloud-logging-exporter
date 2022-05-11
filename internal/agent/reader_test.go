package agent_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nakatamixi/cloud-logging-exporter/internal/agent"
)

func TestReaderAgent_Run(t *testing.T) {
	s := "test"
	mockr := strings.NewReader(s)
	mocke := strings.Builder{}
	a := agent.NewReaderAgent(mockr, &mocke)
	assert.NoError(t, a.Run(context.Background()))
	assert.Equal(t, s, mocke.String())
}
