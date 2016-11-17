package httplog

import (
	"context"
	"net/http"
	"time"
)

type key int

const responseWriterKey key = 0

type Logger interface {
	AfterServeHTTP(*ResponseWriter)
}

func Middleware(logger Logger, next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		rw := &ResponseWriter{
			ResponseWriter: w,
			RequestURI:     r.URL.RequestURI(),
			Start:          now(),
			Done:           now(),
		}
		ctx := newContextWithResponseWriter(r.Context(), rw)

		next.ServeHTTP(rw, r.WithContext(ctx))

		rw.Done = now()
		logger.AfterServeHTTP(rw)
	}
	return http.HandlerFunc(f)
}

func newContextWithResponseWriter(ctx context.Context, rw *ResponseWriter) context.Context {
	return context.WithValue(ctx, responseWriterKey, rw)
}

func WithValue(r *http.Request, key, value interface{}) {
	rw := responseWriterFromContext(r.Context())
	rw.putValue(key, value)
}

func responseWriterFromContext(ctx context.Context) *ResponseWriter {
	return ctx.Value(responseWriterKey).(*ResponseWriter)
}

type ResponseWriter struct {
	http.ResponseWriter

	RequestURI string

	Start time.Time
	Done  time.Time

	Status int
	Size   uint64

	values map[interface{}]interface{}
}

func (r *ResponseWriter) Write(p []byte) (n int, err error) {
	size, err := r.ResponseWriter.Write(p)
	r.Size += uint64(size)
	return size, err
}

func (r *ResponseWriter) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *ResponseWriter) Duration() time.Duration {
	return r.Done.Sub(r.Start)
}

func (r *ResponseWriter) putValue(key, value interface{}) {
	if r.values == nil {
		r.values = map[interface{}]interface{}{}
	}
	r.values[key] = value
}

func (r *ResponseWriter) Value(key interface{}) interface{} {
	return r.values[key]
}

var now func() time.Time = time.Now
