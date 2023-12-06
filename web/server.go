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
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/go-spring-projects/go-spring/web/binding"
)

// A Server defines parameters for running an HTTP server.
type Server struct {
	options  Options
	router   *Router
	renderer Renderer
	httpSvr  *http.Server
}

// NewServer returns a new server instance.
func NewServer(router *Router, options Options) *Server {

	var addr = options.Addr
	if 0 == len(addr) {
		addr = ":8080" // default port: 8080
	}

	var tlsConfig *tls.Config
	if len(options.CertFile) > 0 && len(options.KeyFile) > 0 {
		tlsConfig = &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair(options.CertFile, options.KeyFile)
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
		}
	}

	var jsonRenderer = func(ctx *Context, err error, result interface{}) {

		var code = 0
		var message = ""
		if nil != err {
			var e HttpError
			if errors.As(err, &e) {
				code = e.Code
				message = e.Message
			} else {
				code = http.StatusInternalServerError
				message = err.Error()

				if errors.Is(err, binding.ErrBinding) || errors.Is(err, binding.ErrValidate) {
					code = http.StatusBadRequest
				}
			}
		}

		type response struct {
			Code    int         `json:"code"`
			Message string      `json:"message,omitempty"`
			Data    interface{} `json:"data"`
		}

		ctx.JSON(http.StatusOK, response{Code: code, Message: message, Data: result})
	}

	return &Server{
		options:  options,
		router:   router,
		renderer: RendererFunc(jsonRenderer),
		httpSvr: &http.Server{
			Addr:              addr,
			Handler:           router,
			TLSConfig:         tlsConfig,
			ReadTimeout:       options.ReadTimeout,
			ReadHeaderTimeout: options.ReadHeaderTimeout,
			WriteTimeout:      options.WriteTimeout,
			IdleTimeout:       options.IdleTimeout,
			MaxHeaderBytes:    options.MaxHeaderBytes,
		},
	}
}

// Addr returns the server listen address.
func (s *Server) Addr() string {
	return s.httpSvr.Addr
}

// Run listens on the TCP network address Addr and then
// calls Serve to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
func (s *Server) Run() error {
	if nil != s.httpSvr.TLSConfig {
		return s.httpSvr.ListenAndServeTLS(s.options.CertFile, s.options.KeyFile)
	}
	return s.httpSvr.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the Server's underlying Listener(s).
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSvr.Shutdown(ctx)
}

// Router returns the server router.
func (s *Server) Router() *Router {
	return s.router
}

// NotFound to be used when no route matches.
// This can be used to render your own 404 Not Found errors.
func (s *Server) NotFound(handler http.Handler) {
	s.router.NotFoundHandler = handler
}

// MethodNotAllowed to be used when the request method does not match the route.
// This can be used to render your own 405 Method Not Allowed errors.
func (s *Server) MethodNotAllowed(handler http.Handler) {
	s.router.MethodNotAllowedHandler = handler
}

// Renderer to be used Response renderer in default.
func (s *Server) Renderer(renderer Renderer) {
	s.renderer = renderer
}

// Use appends a MiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify requests and/or responses, and are executed in the order that they are applied to the Router.
func (s *Server) Use(mwf ...MiddlewareFunc) {
	s.router.Use(mwf...)
}

// Match attempts to match the given request against the router's registered routes.
//
// If the request matches a route of this router or one of its subrouters the Route,
// Handler, and Vars fields of the the match argument are filled and this function
// returns true.
//
// If the request does not match any of this router's or its subrouters' routes
// then this function returns false. If available, a reason for the match failure
// will be filled in the match argument's MatchErr field. If the match failure type
// (eg: not found) has a registered handler, the handler is assigned to the Handler
// field of the match argument.
func (s *Server) Match(req *http.Request, match *RouteMatch) bool {
	return s.router.Match(req, match)
}

// ServeHTTP dispatches the handler registered in the matched route.
//
// When there is a match, the route variables can be retrieved calling
// Vars(request).
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// Get returns a route registered with the given name.
func (s *Server) Get(name string) *Route {
	return s.router.Get(name)
}

// StrictSlash defines the trailing slash behavior for new routes. The initial
// value is false.
//
// When true, if the route path is "/path/", accessing "/path" will perform a redirect
// to the former and vice versa. In other words, your application will always
// see the path as specified in the route.
//
// When false, if the route path is "/path", accessing "/path/" will not match
// this route and vice versa.
//
// The re-direct is a HTTP 301 (Moved Permanently). Note that when this is set for
// routes with a non-idempotent method (e.g. POST, PUT), the subsequent re-directed
// request will be made as a GET by most clients. Use middleware or client settings
// to modify this behaviour as needed.
//
// Special case: when a route sets a path prefix using the PathPrefix() method,
// strict slash is ignored for that route because the redirect behavior can't
// be determined from a prefix alone. However, any subrouters created from that
// route inherit the original StrictSlash setting.
func (s *Server) StrictSlash(value bool) {
	s.router.StrictSlash(value)
}

// SkipClean defines the path cleaning behaviour for new routes. The initial
// value is false. Users should be careful about which routes are not cleaned
//
// When true, if the route path is "/path//to", it will remain with the double
// slash. This is helpful if you have a route like: /fetch/http://xkcd.com/534/
//
// When false, the path will be cleaned, so /fetch/http://xkcd.com/534/ will
// become /fetch/http/xkcd.com/534
func (s *Server) SkipClean(value bool) {
	s.router.SkipClean(value)
}

// UseEncodedPath tells the router to match the encoded original path
// to the routes.
// For eg. "/path/foo%2Fbar/to" will match the path "/path/{var}/to".
//
// If not called, the router will match the unencoded path to the routes.
// For eg. "/path/foo%2Fbar/to" will match the path "/path/foo/bar/to"
func (s *Server) UseEncodedPath() {
	s.router.UseEncodedPath()
}

// NewRoute registers an empty route.
func (s *Server) NewRoute() *Route {
	return s.router.NewRoute()
}

// Name registers a new route with a name.
// See Route.Name().
func (s *Server) Name(name string) *Route {
	return s.router.Name(name)
}

// Handle registers a new route with a matcher for the URL path.
// See Route.Path() and Route.Handler().
func (s *Server) Handle(path string, handler http.Handler) *Route {
	return s.router.Handle(path, handler)
}

// HandleFunc registers a new route with a matcher for the URL path.
// See Route.Path() and Route.HandlerFunc().
func (s *Server) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route {
	return s.router.HandleFunc(path, f)
}

// Bind registers a new route with a matcher for the URL path.
//
// func(ctx context.Context)
//
// func(ctx context.Context) R
//
// func(ctx context.Context, req T) R
//
// func(ctx context.Context, req T) (R, error)
func (s *Server) Bind(path string, f interface{}, r ...Renderer) *Route {
	var renderer = s.renderer
	if len(r) > 0 {
		renderer = r[0]
	}
	return s.Handle(path, Bind(f, renderer))
}

// Headers registers a new route with a matcher for request header values.
// See Route.Headers().
func (s *Server) Headers(pairs ...string) *Route {
	return s.router.Headers(pairs...)
}

// MatcherFunc registers a new route with a custom matcher function.
// See Route.MatcherFunc().
func (s *Server) MatcherFunc(f MatcherFunc) *Route {
	return s.router.MatcherFunc(f)
}

// Methods registers a new route with a matcher for HTTP methods.
// See Route.Methods().
func (s *Server) Methods(methods ...string) *Route {
	return s.router.Methods(methods...)
}

// Path registers a new route with a matcher for the URL path.
// See Route.Path().
func (s *Server) Path(tpl string) *Route {
	return s.router.Path(tpl)
}

// PathPrefix registers a new route with a matcher for the URL path prefix.
// See Route.PathPrefix().
func (s *Server) PathPrefix(tpl string) *Route {
	return s.router.PathPrefix(tpl)
}

// Queries registers a new route with a matcher for URL query values.
// See Route.Queries().
func (s *Server) Queries(pairs ...string) *Route {
	return s.router.Queries(pairs...)
}

// Schemes registers a new route with a matcher for URL schemes.
// See Route.Schemes().
func (s *Server) Schemes(schemes ...string) *Route {
	return s.router.Schemes(schemes...)
}

// BuildVarsFunc registers a new route with a custom function for modifying
// route variables before building a URL.
func (s *Server) BuildVarsFunc(f BuildVarsFunc) *Route {
	return s.router.BuildVarsFunc(f)
}

// Walk walks the router and all its sub-routers, calling walkFn for each route
// in the tree. The routes are walked in the order they were added. Sub-routers
// are explored depth-first.
func (s *Server) Walk(walkFn WalkFunc) error {
	return s.router.Walk(walkFn)
}
