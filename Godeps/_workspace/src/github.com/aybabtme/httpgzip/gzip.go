// Package httpgzip implements a http.Handler that compresses all HTTP
// operations with GZIP if the client supports it.  It supports
// long-lived operations such as streams of data over HTTP, or short,
// regular HTTP operations.
//
//    gzh := httpgzip.NewHandler(myHandler)
//    log.Fatal(http.ListenAndServe(":8080", gzh))
//
package httpgzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Compile checks for interface implementations.
var (
	_ http.Handler        = &GzipHandler{}
	_ http.Flusher        = &gzipResponseFlusher{}
	_ http.ResponseWriter = &gzipResponseFlusher{}
	_ http.ResponseWriter = &gzipResponseWriter{}
)

type gzipResponseFlusher struct {
	io.Writer
	http.ResponseWriter
	http.Flusher
	gziper *gzip.Writer
}

func (w gzipResponseFlusher) Write(b []byte) (int, error) {
	if "" == w.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

func (w *gzipResponseFlusher) Flush() {
	_ = w.gziper.Flush()
	w.Flusher.Flush()
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gziper *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if "" == w.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.gziper.Write(b)
}

// GzipHandler wraps a handler.  Before passing a HTTP transaction to
// the underlying handler, it will try to compress the response writer.
//
// This handler supports both short lived and long lived HTTP, which makes
// it appropriate to serve streamed data.
type GzipHandler struct {
	wrap http.Handler

	// OnError is a hook you can provide to GzipHandler if you want to
	// handle cases where the handler can't process a request or has
	// encountered an error while processing it.
	//
	// If not set, will default to DefaultOnError
	OnError func(err error, w http.ResponseWriter, r *http.Request)
}

// NewHandler wraps h with a GzipHandler.
func NewHandler(h http.Handler) *GzipHandler {
	return &GzipHandler{
		wrap:    h,
		OnError: DefaultOnError,
	}
}

// DefaultOnError logs errors to the standard log.Printf.
func DefaultOnError(err error, w http.ResponseWriter, r *http.Request) {
	log.Printf("[httpgzip] %v", err)
}

func (g *GzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		g.wrap.ServeHTTP(w, r)
		return
	}

	w.Header().Add("Content-Encoding", "gzip")

	gz := gzip.NewWriter(w)
	defer func() {
		if err := gz.Close(); err != nil {
			g.OnError(fmt.Errorf("closing gzip writer, %v", err), w, r)
		}
	}()

	flush, ok := w.(http.Flusher)
	if !ok {
		g.OnError(fmt.Errorf("writer to %q is not a http.Flusher", r.RemoteAddr), w, r)

		gzr := &gzipResponseWriter{
			ResponseWriter: w,
			gziper:         gz,
		}
		g.wrap.ServeHTTP(gzr, r)
		return
	}

	gzr := &gzipResponseFlusher{
		Writer:         gz,
		ResponseWriter: w,
		Flusher:        flush,
		gziper:         gz,
	}
	g.wrap.ServeHTTP(gzr, r)
}
