package httplog

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

type MockHTTPResponseWriter struct {
	N   int
	Err error
}

func (w *MockHTTPResponseWriter) Write(p []byte) (int, error) {
	return w.N, w.Err
}

func (w *MockHTTPResponseWriter) WriteHeader(status int) {
}

func (w *MockHTTPResponseWriter) Header() http.Header {
	return nil
}

type MockLogger struct {
	t *testing.T

	Request *http.Request

	Called bool

	RequestURI string

	Start time.Time
	Done  time.Time

	Status int
	Size   uint64

	Values map[interface{}]interface{}
}

func (l *MockLogger) AfterServeHTTP(rw *ResponseWriter) {
	if l.Called {
		l.t.Error("should not have been called prior")
	}

	l.Called = true

	if rw.Request != l.Request {
		l.t.Errorf("%v != %v", rw.Request, l.Request)
	}
	if rw.RequestURI != l.RequestURI {
		l.t.Errorf("%v != %v", rw.RequestURI, l.RequestURI)
	}
	if rw.Start != l.Start {
		l.t.Errorf("%v != %v", rw.Start, l.Start)
	}
	if rw.Done != l.Done {
		l.t.Errorf("%v != %v", rw.Done, l.Done)
	}
	if rw.Status != l.Status {
		l.t.Errorf("%v != %v", rw.Status, l.Status)
	}
	if rw.Size != l.Size {
		l.t.Errorf("%v != %v", rw.Size, l.Size)
	}
	if !reflect.DeepEqual(rw.values, l.Values) {
		l.t.Errorf("%v != %v", rw.values, l.Values)
	}
}

func (l *MockLogger) AssertCalled() {
	if !l.Called {
		l.t.Error("should have been called")
	}
}
