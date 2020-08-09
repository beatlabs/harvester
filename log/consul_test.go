package log

import (
	"bytes"
	"log"
	"os"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestConsul(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	logger := Consul()

	t.Run("Level", func(t *testing.T) {
		logger.Log(hclog.Warn, "TEST: %d", 123)
		assert.Contains(t, buf.String(), "WARN: TEST: [123]")
		buf.Reset()
	})

	t.Run("Trace", func(t *testing.T) {
		logger.Trace("TEST: %d", 123)
		assert.Contains(t, buf.String(), "DEBUG: TEST: [123]")
		buf.Reset()
	})

	t.Run("Debug", func(t *testing.T) {
		logger.Debug("TEST: %d", 123)
		assert.Contains(t, buf.String(), "DEBUG: TEST: [123]")
		buf.Reset()
	})

	t.Run("Info", func(t *testing.T) {
		logger.Info("TEST: %d", 123)
		assert.Contains(t, buf.String(), "INFO: TEST: [123]")
		buf.Reset()
	})

	t.Run("Warn", func(t *testing.T) {
		logger.Warn("TEST: %d", 123)
		assert.Contains(t, buf.String(), "WARN: TEST: [123]")
		buf.Reset()
	})

	t.Run("Error", func(t *testing.T) {
		logger.Error("TEST: %d", 123)
		assert.Contains(t, buf.String(), "ERROR: TEST: [123]")
		buf.Reset()
	})

	assert.True(t, logger.IsTrace())
	assert.True(t, logger.IsDebug())
	assert.True(t, logger.IsInfo())
	assert.True(t, logger.IsWarn())
	assert.True(t, logger.IsError())
	assert.Empty(t, logger.ImpliedArgs())
	assert.Equal(t, logger, logger.With())
	assert.Equal(t, "consul", logger.Name())
	assert.Equal(t, logger, logger.Named("test"))
	assert.Equal(t, logger, logger.ResetNamed("test"))
	assert.IsType(t, &log.Logger{}, logger.StandardLogger(nil))
	assert.Equal(t, os.Stderr, logger.StandardWriter(nil))
}
