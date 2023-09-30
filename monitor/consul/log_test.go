package consul

import (
	"bytes"
	"log/slog"
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

	logger.Log(hclog.Off, "Name: %s", "One")

	assert.Empty(t, b.String())

	b.Reset()

	logger.Log(hclog.NoLevel, "Name: %s", "One")

	assert.Empty(t, b.String())

	b.Reset()

	logger.Log(hclog.Trace, "Name: %s", "One")

	assert.Empty(t, b.String())

	b.Reset()

	logger.Log(hclog.Debug, "Name: %s", "One")

	assert.Empty(t, b.String())

	b.Reset()

	logger.Log(hclog.Info, "Name: %s", "One")

	assert.Contains(t, b.String(), `level=INFO msg="Name: One"`)

	b.Reset()

	logger.Log(hclog.Warn, "Name: %s", "One")

	assert.Contains(t, b.String(), `level=WARN msg="Name: One"`)

	b.Reset()

	logger.Log(hclog.Error, "Name: %s", "One")

	assert.Contains(t, b.String(), `level=ERROR msg="Name: One"`)

	b.Reset()
}
