package httplog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	NanosPerMicros = 1000000.0
)

func Middleware(h http.Handler) http.Handler {
	l := &Logger{}
	return l.Middleware(h)
}

type ContextCreator func(w http.ResponseWriter, r *http.Request) *Context

type ContextFormatter func(*Context) string

type Logger struct {
	Creator   ContextCreator
	Formatter ContextFormatter
	io.Writer
}

func (l *Logger) Middleware(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		c := l.newContext(w, r)
		h.ServeHTTP(c, r)
		c.update()
		l.writeContext(c)
	}
	return http.HandlerFunc(f)
}

func (l *Logger) newContext(w http.ResponseWriter, r *http.Request) *Context {
	if l.Creator != nil {
		return l.Creator(w, r)
	}
	return NewContext(w, r)
}

func (l *Logger) writeContext(c *Context) {
	fmt.Fprintln(l.getWriter(), l.getResult(c))
}

func (l *Logger) getWriter() io.Writer {
	if l.Writer != nil {
		return l.Writer
	}
	return os.Stdout
}

func (l *Logger) getResult(c *Context) string {
	if l.Formatter != nil {
		return l.Formatter(c)
	}
	return "NEED TO IMPLEMENT"
}

type Context struct {
	http.ResponseWriter

	Request   *http.Request
	Path      string
	Ident     string
	User      string
	TimeStart time.Time
	TimeDone  time.Time
	Status    int
	Size      int
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		ResponseWriter: w,
		Request:        r,
		Path:           r.URL.Path,
		TimeStart:      time.Now(),
		TimeDone:       time.Now(),
	}
}

func (c *Context) Write(data []byte) (int, error) {
	size, err := c.ResponseWriter.Write(data)
	c.Size += size
	return size, err
}

func (c *Context) WriteHeader(status int) {
	c.Status = status
	c.ResponseWriter.WriteHeader(c.Status)
}

func (c *Context) update() {
	c.TimeDone = time.Now()
	if c.Status == 0 {
		c.Status = http.StatusOK
	}
}
