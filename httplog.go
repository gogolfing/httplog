package httplog

import (
	"fmt"
	"net/http"
	"time"
)

const (
	NANOS_PER_MICROS = 1000000.0
)

type context struct {
	method    string
	path      string
	startTime time.Time
	status    int
	size      int
	http.ResponseWriter
}

func (c *context) Write(data []byte) (int, error) {
	c.size += len(data)
	return c.ResponseWriter.Write(data)
}

func (c *context) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

func Middleware(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		c := &context{
			r.Method,
			r.URL.Path,
			time.Now(),
			0,
			0,
			w,
		}
		h.ServeHTTP(c, r)
		c.finish()
	}
	return http.HandlerFunc(f)
}

func (c *context) finish() {
	c.writeLog()
	c.ResponseWriter = nil
}

func (c *context) writeLog() {
	if c.status == 0 {
		c.status = 200
	}
	ms := float64(time.Since(c.startTime).Nanoseconds()) / NANOS_PER_MICROS
	fmt.Printf("%v %v %v %vB %.4fms\n", c.method, c.path, c.status, c.size, ms)
}
