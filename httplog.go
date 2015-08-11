package httplog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
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

type Printer interface {
	Println(...interface{})
}

type WriterPrinter struct {
	*sync.Mutex
	io.Writer
}

func NewWriterPrinter(w io.Writer) Printer {
	return &WriterPrinter{
		&sync.Mutex{},
		w,
	}
}

func (wp *WriterPrinter) Println(v ...interface{}) {
	wp.Lock()
	defer wp.Unlock()
	fmt.Fprintln(wp, v...)
}

type Logger struct {
	Creator   ContextCreator
	Formatter ContextFormatter
	Printer
}

func NewLogger(c ContextCreator, f ContextFormatter, p Printer) *Logger {
	return &Logger{
		Creator:   c,
		Formatter: f,
		Printer:   p,
	}
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
	if l.Printer == nil {
		l.Printer = NewWriterPrinter(os.Stdout)
	}
	l.Println(l.getResult(c))
}

func (l *Logger) getResult(c *Context) string {
	if l.Formatter != nil {
		return l.Formatter(c)
	}
	return FormatContext(c)
}

type Context struct {
	http.ResponseWriter

	Request    *http.Request
	RequestURI string
	Identity   string
	AuthUser   string
	TimeStart  time.Time
	TimeDone   time.Time
	Status     int
	Size       int
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		ResponseWriter: w,
		Request:        r,
		RequestURI:     r.URL.RequestURI(),
		TimeStart:      time.Now(),
		TimeDone:       time.Now(),
	}
}

func FormatContext(c *Context) string {
	ms := float64(c.TimeDone.Sub(c.TimeStart).Nanoseconds()) / NanosPerMicros
	return fmt.Sprintf("%v %v %v %vB %.4fms\n", c.Request.Method, c.RequestURI, c.Status, c.Size, ms)
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
