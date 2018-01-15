//Package httplog provides a Middleware function along with an http.ResponseWriter
//implementation that records response data for processing after a Request is complete.
package httplog

import (
	"context"
	"net/http"
	"time"
)

type key int

const responseWriterKey key = 0

//Logger is the interface required to receive response information after an
//http.Handler has been served.
type Logger interface {
	//Called from Middleware() after next.ServeHTTP() has returned.
	AfterServeHTTP(*ResponseWriter)
}

//Middleware creates a new http.Handler that calls next with a ResponseWriter.
//Once next.ServeHTTP() has returned, final statistics are collected, and
//logger.AfterServeHTTP() is called.
//
//Normally in the Request, Context flow of data, only downstream http.Handlers are able
//to know about values set in the Context from prior middleware.
//If you desire to have this logger middleware be at the outermost level, then
//you will need to use WithValue() to set values within a ResponseWriter to
//store those values for processing after handling the Request.
func Middleware(logger Logger, next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		rw := &ResponseWriter{
			ResponseWriter: w,
			Request:        r,
			RequestURI:     r.URL.RequestURI(),
			Start:          now(),
			Done:           now(),
		}
		ctx := newContextWithResponseWriter(r.Context(), rw)

		next.ServeHTTP(rw, r.WithContext(ctx))

		rw.Done = now()

		if rw.Status == 0 {
			rw.Status = http.StatusOK
		}

		logger.AfterServeHTTP(rw)
	}
	return http.HandlerFunc(f)
}

func newContextWithResponseWriter(ctx context.Context, rw *ResponseWriter) context.Context {
	return context.WithValue(ctx, responseWriterKey, rw)
}

//WithValue associates key and value in r's ResponseWriter that can later be
//retrieved with ResponseWriter.Value(key).
//
//This allows downstream handlers from this logger middleware to set values
//on the ResponseWriter for processing after a Request has been served.
func WithValue(r *http.Request, key, value interface{}) {
	rw := responseWriterFromContext(r.Context())
	rw.putValue(key, value)
}

func responseWriterFromContext(ctx context.Context) *ResponseWriter {
	return ctx.Value(responseWriterKey).(*ResponseWriter)
}

//ResponseWriter is an http.ResponseWriter implementation that records status,
//size, start time, and completed time for the execution of an http.Request.
//
//Instances of ResponseWriter are passed into next.ServeHTTP() from Middleware()
//in order to record response statistics that can be examined and processed
//after the Request has completed.
type ResponseWriter struct {
	//ResponseWriter is the promoted http.ResponseWriter implementation.
	http.ResponseWriter

	//Request is the Request that the Handler is serving.
	Request *http.Request

	//RequestURI is set before next.ServeHTTP() in Middleware() so that the full
	//URL is recorded before downstream http.Handlers possibly modify the Request URL.
	RequestURI string

	//Start is set to time.Now() when Middleware() receives a Request.
	Start time.Time
	//Done is set to time.Now() when Middleware()'s next.ServeHTTP() returns.
	Done time.Time

	//Status is the recorded status code from http.ResponseWriter.WriteHeader().
	Status int
	//Size is the total number of bytes written to http.ResponseWriter.Write()
	//throughout the lifetime of all downstream http.Handlers.
	Size uint64

	values map[interface{}]interface{}
}

//Write is the http.ResponseWriter implementation of Write that records the number
//bytes written and passes p along to r.ResponseWriter.
func (r *ResponseWriter) Write(p []byte) (n int, err error) {
	size, err := r.ResponseWriter.Write(p)
	r.Size += uint64(size)
	return size, err
}

//WriteHeader is the http.ResponseWriter implementation of WriteHeader that records
//status and passes status along to r.ResponseWriter.
func (r *ResponseWriter) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

//Duration returns the length of time that all downstream http.Handlers took to
//execute.
func (r *ResponseWriter) Duration() time.Duration {
	return r.Done.Sub(r.Start)
}

func (r *ResponseWriter) putValue(key, value interface{}) {
	if r.values == nil {
		r.values = map[interface{}]interface{}{}
	}
	r.values[key] = value
}

//Value returns the value associated with key, for a given Request's ResponseWriter,
//that was previously set by WithValue().
func (r *ResponseWriter) Value(key interface{}) interface{} {
	return r.values[key]
}

var now func() time.Time = time.Now
