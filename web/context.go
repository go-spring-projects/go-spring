/*
 * Copyright 2023 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package web

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/go-spring-projects/go-spring/web/binding"
	"github.com/go-spring-projects/go-spring/web/render"
	"github.com/gorilla/mux"
)

type contextKey struct{}

func WithContext(parent context.Context, ctx *Context) context.Context {
	return context.WithValue(parent, contextKey{}, ctx)
}

func FromContext(ctx context.Context) *Context {
	if v := ctx.Value(contextKey{}); v != nil {
		return v.(*Context)
	}
	return nil
}

type Context struct {
	// A ResponseWriter interface is used by an HTTP handler to
	// construct an HTTP response.
	Writer http.ResponseWriter

	// A Request represents an HTTP request received by a server
	// or to be sent by a client.
	Request *http.Request

	// SameSite allows a server to define a cookie attribute making it impossible for
	// the browser to send this cookie along with cross-site requests.
	sameSite http.SameSite
}

// Context returns the request's context.
func (c *Context) Context() context.Context {
	return c.Request.Context()
}

// ContentType returns the request header `Content-Type`.
func (c *Context) ContentType() string {
	contentType := c.Request.Header.Get("Content-Type")
	return contentType
}

// Header returns the named header in the request.
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// Cookie returns the named cookie provided in the request.
func (c *Context) Cookie(name string) string {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return ""
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val
}

// PathParam returns the named variables in the request.
func (c *Context) PathParam(name string) string {
	if params := mux.Vars(c.Request); nil != params {
		if value, ok := params[name]; ok {
			return value
		}
	}
	return ""
}

// QueryParam returns the named query in the request.
func (c *Context) QueryParam(name string) string {
	if values := c.Request.URL.Query(); nil != values {
		if value, ok := values[name]; ok && len(value) > 0 {
			return value[0]
		}
	}
	return ""
}

// FormParams returns the form in the request.
func (c *Context) FormParams() (url.Values, error) {
	if err := c.Request.ParseForm(); nil != err {
		return nil, err
	}
	return c.Request.Form, nil
}

// MultipartParams returns a request body as multipart/form-data.
// The whole request body is parsed and up to a total of maxMemory bytes of its file parts are stored in memory, with the remainder stored on disk in temporary files.
func (c *Context) MultipartParams(maxMemory int64) (*multipart.Form, error) {
	if !strings.Contains(c.ContentType(), binding.MIMEMultipartForm) {
		return nil, fmt.Errorf("require `multipart/form-data` request")
	}

	if nil == c.Request.MultipartForm {
		if err := c.Request.ParseMultipartForm(maxMemory); nil != err {
			return nil, err
		}
	}
	return c.Request.MultipartForm, nil
}

// RequestBody returns the request body.
func (c *Context) RequestBody() io.Reader {
	return c.Request.Body
}

// IsWebsocket returns true if the request headers indicate that a websocket
// handshake is being initiated by the client.
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.Request.Header.Get("Connection")), "upgrade") &&
		strings.EqualFold(c.Request.Header.Get("Upgrade"), "websocket") {
		return true
	}
	return false
}

// SetSameSite with cookie
func (c *Context) SetSameSite(samesite http.SameSite) {
	c.sameSite = samesite
}

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// SetHeader is an intelligent shortcut for c.Writer.Header().Set(key, value).
// It writes a header in the response.
// If value == "", this method removes the header `c.Writer.Header().Del(key)`
func (c *Context) SetHeader(key, value string) {
	if value == "" {
		c.Writer.Header().Del(key)
		return
	}
	c.Writer.Header().Set(key, value)
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, render render.Renderer) error {
	if code > 0 {
		if len(c.Writer.Header().Get("Content-Type")) <= 0 {
			if contentType := render.ContentType(); len(contentType) > 0 {
				c.Writer.Header().Set("Content-Type", contentType)
			}
		}
		c.Writer.WriteHeader(code)
	}
	return render.Render(c.Writer)
}

// Redirect returns an HTTP redirect to the specific location.
func (c *Context) Redirect(code int, location string) error {
	return c.Render(-1, render.RedirectRenderer{Code: code, Request: c.Request, Location: location})
}

// String writes the given string into the response body.
func (c *Context) String(code int, format string, args ...interface{}) error {
	return c.Render(code, render.TextRenderer{Format: format, Args: args})
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(code int, contentType string, data []byte) error {
	return c.Render(code, render.BinaryRenderer{DataType: contentType, Data: data})
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj interface{}) error {
	return c.Render(code, render.JsonRenderer{Data: obj})
}

// IndentedJSON serializes the given struct as pretty JSON (indented + endlines) into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) IndentedJSON(code int, obj interface{}) error {
	return c.Render(code, render.JsonRenderer{Data: obj, Indent: "  "})
}

// XML serializes the given struct as XML into the response body.
// It also sets the Content-Type as "application/xml".
func (c *Context) XML(code int, obj interface{}) error {
	return c.Render(code, render.XmlRenderer{Data: obj})
}

// IndentedXML serializes the given struct as pretty XML (indented + endlines) into the response body.
// It also sets the Content-Type as "application/xml".
func (c *Context) IndentedXML(code int, obj interface{}) error {
	return c.Render(code, render.XmlRenderer{Data: obj, Indent: "  "})
}

// File writes the specified file into the body stream in an efficient way.
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// FileAttachment writes the specified file into the body stream in an efficient way
// On the client side, the file will typically be downloaded with the given filename
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+escapeQuotes(filename)+`"`)
	} else {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.Writer, c.Request, filepath)
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func (c *Context) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// ClientIP implements one best effort algorithm to return the real client IP.
// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
// If it is it will then try to parse the headers defined in RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]).
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP (coming from Request.RemoteAddr) is returned.
func (c *Context) ClientIP() string {
	// It also checks if the remoteIP is a trusted proxy or not.
	// In order to perform this validation, it will see if the IP is contained within at least one of the CIDR blocks
	// defined by Engine.SetTrustedProxies()
	remoteIP := net.ParseIP(c.RemoteIP())
	if remoteIP == nil {
		return ""
	}

	for _, headerName := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		if ns := strings.Split(c.Request.Header.Get(headerName), ","); len(ns) > 0 && len(ns[0]) > 0 {
			return ns[0]
		}
	}
	return remoteIP.String()
}

// https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}
