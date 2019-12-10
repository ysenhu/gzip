package gzip

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newGinInstance(payload []byte, middleware ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.Use(middleware...)

	g.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf8", payload)
	})

	return g
}

type NopWriter struct {
	header http.Header
}

func NewNopWriter() *NopWriter {
	return &NopWriter{
		header: make(http.Header),
	}
}

func (n *NopWriter) Header() http.Header {
	return n.header
}

func (n *NopWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (n *NopWriter) WriteHeader(_ int) {
	// relax
}

func TestNewHandler_Checks(t *testing.T) {
	assert.NotPanics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: 100,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: -3,
			MinContentLength: 100,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: 10,
			MinContentLength: 100,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: 0,
		})
	})

	assert.Panics(t, func() {
		NewHandler(Config{
			CompressionLevel: 5,
			MinContentLength: -1,
		})
	})
}

func BenchmarkSoleGin_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload)
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_SmallPayload(b *testing.B) {
	var (
		g = newGinInstance(smallPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkSoleGin_BigPayload(b *testing.B) {
	var (
		g = newGinInstance(bigPayload)
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "" {
		b.Fatalf("Content-Encoding is not empty, but %s", encoding)
	}
}

func BenchmarkGinWithDefaultHandler_BigPayload(b *testing.B) {
	var (
		g = newGinInstance(bigPayload, DefaultHandler().Gin)
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		w = NewNopWriter()
	)

	r.Header.Set("Accept-Encoding", "gzip")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.ServeHTTP(w, r)
	}

	b.StopTimer()
	if encoding := w.Header().Get("Content-Encoding"); encoding != "gzip" {
		b.Fatalf("Content-Encoding is not gzip, but %q", encoding)
	}
}