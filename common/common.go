package common

import (
	"net/http"

	"github.com/gogolfing/httplog"
)

type Indentity func(r *http.Request) string

type AuthUser func(r *http.Request) string

func EmptyIdentity(_ *http.Request) string {
	return ""
}

func EmptyAuthUser(_ *http.Request) string {
	return ""
}

func EmptyCreator(w http.ResponseWriter, r *http.Request) *httplog.Context {
	c := httplog.NewContext(w, r)
	c.Identity = EmptyIdentity(r)
	c.AuthUser = EmptyAuthUser(r)
	return c
}

func NewEmptyLogger(printer httplog.Printer) *httplog.Logger {
	return httplog.NewLogger(EmptyCreator, FormatContext, printer)
}

func NewCreator(i Indentity, au AuthUser) httplog.ContextCreator {
	f := func(w http.ResponseWriter, r *http.Request) *httplog.Context {
		c := httplog.NewContext(w, r)
		c.Identity = i(r)
		c.AuthUser = au(r)
		return c
	}
	return f
}

func FormatContext(c *httplog.Context) string {
	return "this is supposed to be common log format"
}
