package httplog

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var TimeOne = time.Now()

func TestMiddleware_SetStatus(t *testing.T) {
	oldNow := now
	defer func() {
		now = oldNow
	}()
	now = func() time.Time {
		return TimeOne
	}

	r := httptest.NewRequest("", "https://www.example.com/path", nil)

	logger := &MockLogger{
		t:          t,
		Request:    r,
		RequestURI: "/path",
		Start:      TimeOne,
		Done:       TimeOne,
		Status:     http.StatusOK,
		Size:       uint64(len("finalHandler")),
		Values: map[interface{}]interface{}{
			1: "one",
		},
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//notice that we are testing WithValue() here as well.
		WithValue(r, 1, "one")

		fmt.Fprint(w, "finalHandler")
	})

	handler := Middleware(logger, finalHandler)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	logger.AssertCalled()
}

func TestResponseWriter_Write(t *testing.T) {
	wantErr := fmt.Errorf("error")
	rw := &ResponseWriter{
		ResponseWriter: &MockHTTPResponseWriter{
			N:   10,
			Err: wantErr,
		},
	}
	n, err := rw.Write([]byte{1, 2, 3, 4})

	if n != 10 || err != wantErr {
		t.Fatal()
	}
	if rw.Size != 10 {
		t.Fatal()
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rw := &ResponseWriter{
		ResponseWriter: &MockHTTPResponseWriter{},
	}

	rw.WriteHeader(-1)

	if rw.Status != -1 {
		t.Fatal()
	}
}

func TestReponseWriter_IsAHijacker(t *testing.T) {
	func(h http.Hijacker) {
	}(&ResponseWriter{})
}

func TestResponseWriter_Duration(t *testing.T) {
	start := time.Now()
	done := start.Add(time.Duration(10))
	rw := &ResponseWriter{
		Start: start,
		Done:  done,
	}

	if time.Duration(10) != rw.Duration() {
		t.Fatal()
	}
}

func TestResponseWriter_putValue(t *testing.T) {
	rw := &ResponseWriter{}

	rw.putValue(1, "one")

	if rw.values[1] != "one" {
		t.Fatal()
	}
}

func TestResponseWriter_Value(t *testing.T) {
	rw := &ResponseWriter{}

	rw.putValue(1, "one")

	if rw.Value(1) != "one" {
		t.Fatal()
	}
}
