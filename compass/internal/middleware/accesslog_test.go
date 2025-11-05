package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	requestid "github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	// Copy record to make attrs accessible later
	copied := slog.Record{Time: r.Time, Message: r.Message, Level: r.Level, PC: r.PC}
	r.Attrs(func(a slog.Attr) bool {
		copied.AddAttrs(a)
		return true
	})
	h.mu.Lock()
	h.records = append(h.records, copied)
	h.mu.Unlock()
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For tests, keep it simple: attach attrs by appending during Handle.
	return h
}

func (h *captureHandler) WithGroup(name string) slog.Handler { return h }

func TestAccessLogger_EmitsRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ch := &captureHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(ch))
	t.Cleanup(func() { slog.SetDefault(prev) })

	r := gin.New()
	r.Use(requestid.New(), AccessLogger())
	r.GET("/hello", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("X-Request-ID", "accesslog-test")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ch.mu.Lock()
	defer ch.mu.Unlock()
	require.NotEmpty(t, ch.records, "expected at least one log record")

	// Find the http_request record
	var found *slog.Record
	for i := range ch.records {
		if ch.records[i].Message == "http_request" {
			found = &ch.records[i]
			break
		}
	}
	require.NotNil(t, found, "expected http_request log record, got %d records", len(ch.records))

	// Extract attrs to a map for assertions
	got := map[string]any{}
	found.Attrs(func(a slog.Attr) bool { got[a.Key] = a.Value.Any(); return true })

	assert.Equal(t, "accesslog-test", got["request_id"], "request_id mismatch")
	assert.Equal(t, http.MethodGet, got["method"], "method mismatch")
	assert.Equal(t, "/hello", got["path"], "path mismatch")
	_, ok := got["status"]
	assert.True(t, ok, "missing status attr")
}
