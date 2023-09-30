package consul

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func Test_clog(t *testing.T) {
	defaultLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(defaultLogger)
	})

	var b bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&b, nil)))

	t.Run("configs", func(t *testing.T) {
		assert.True(t, logger.IsTrace())
		assert.True(t, logger.IsDebug())
		assert.True(t, logger.IsInfo())
		assert.True(t, logger.IsWarn())
		assert.True(t, logger.IsError())
		assert.Empty(t, logger.ImpliedArgs())
		assert.NotNil(t, logger.With(nil))
		assert.Equal(t, "consul", logger.Name())
		assert.NotNil(t, logger.Named(""))
		assert.NotNil(t, logger.ResetNamed(""))
		assert.NotNil(t, logger.StandardLogger(nil))
		assert.Equal(t, os.Stderr, logger.StandardWriter(nil))
		assert.Equal(t, hclog.NoLevel, logger.GetLevel())
	})

	t.Run("off", func(t *testing.T) {
		logger.Log(hclog.Off, "Name: %s", "One")
		assert.Empty(t, b.String())
		b.Reset()
	})

	t.Run("off", func(t *testing.T) {
		logger.Log(hclog.NoLevel, "Name: %s", "One")
		assert.Empty(t, b.String())
		b.Reset()
	})

	t.Run("trace", func(t *testing.T) {
		logger.Log(hclog.Trace, "Name: %s", "One")
		assert.Empty(t, b.String())
		b.Reset()
	})

	t.Run("debug", func(t *testing.T) {
		logger.Log(hclog.Debug, "Name: %s", "One")
		assert.Empty(t, b.String())
		b.Reset()
	})

	t.Run("info", func(t *testing.T) {
		logger.Log(hclog.Info, "Name: %s", "One")
		assert.Contains(t, b.String(), `level=INFO msg="Name: One"`)
		b.Reset()
	})

	t.Run("warn", func(t *testing.T) {
		logger.Log(hclog.Warn, "Name: %s", "One")
		assert.Contains(t, b.String(), `level=WARN msg="Name: One"`)
		b.Reset()
	})

	t.Run("error", func(t *testing.T) {
		logger.Log(hclog.Error, "Name: %s", "One")
		assert.Contains(t, b.String(), `level=ERROR msg="Name: One"`)
		b.Reset()
	})
}
