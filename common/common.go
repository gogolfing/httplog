package common

import (
	"fmt"
	"net/http"

	"github.com/gogolfing/httplog"
)

const (
	CommonLogDateFormat = "02/Jan/2006:15:04:05 -0700"
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

func CombinedFormatContext(c *httplog.Context) string {
	return fmt.Sprintf(`%s "%s" "%s"`,
		FormatContext(c),
		c.Request.Header.Get("Referer"),
		c.Request.Header.Get("User-Agent"),
	)
}

func FormatContext(c *httplog.Context) string {
	identity := c.Identity
	if len(identity) == 0 {
		identity = "-"
	}
	authUser := c.AuthUser
	if len(authUser) == 0 {
		authUser = "-"
	}
	uri := c.Path
	if len(c.Request.URL.RawQuery) > 0 {
		uri += "?" + c.Request.URL.RawQuery
	}
	return fmt.Sprintf(`%s %s %s [%s] "%s %s %s" %d %d`,
		c.Request.RemoteAddr,
		identity,
		authUser,
		c.TimeDone.Format(CommonLogDateFormat),
		c.Request.Method,
		uri,
		c.Request.Proto,
		c.Status,
		c.Size,
	)
}
