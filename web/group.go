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
	"net/http"

	"github.com/go-spring-projects/go-spring/web/internal/mux"
)

// MiddlewareFunc is a function which receives an http.Handler and returns another http.Handler.
type MiddlewareFunc = mux.MiddlewareFunc

type RouterGroup interface {
	// Handler dispatches the handler registered in the matched route.
	http.Handler

	// BasePath returns the base path of router group.
	BasePath() string

	// Use appends a MiddlewareFunc to the chain.
	Use(mwf ...MiddlewareFunc)
	// Renderer to be used Response renderer in default.
	Renderer(renderer Renderer)

	// NotFound to be used when no route matches.
	NotFound(handler http.Handler)
	// MethodNotAllowed to be used when the request method does not match the route.
	MethodNotAllowed(handler http.Handler)

	// StrictSlash defines the trailing slash behavior for new routes. The initial
	// value is false.
	StrictSlash(value bool)
	// SkipClean defines the path cleaning behaviour for new routes. The initial
	// value is false. Users should be careful about which routes are not cleaned
	//
	// When true, if the route path is "/path//to", it will remain with the double
	// slash. This is helpful if you have a route like: /fetch/http://xkcd.com/534/
	//
	// When false, the path will be cleaned, so /fetch/http://xkcd.com/534/ will
	// become /fetch/http/xkcd.com/534
	SkipClean(value bool)
	// UseEncodedPath tells the router to match the encoded original path
	// to the routes.
	// For eg. "/path/foo%2Fbar/to" will match the path "/path/{var}/to".
	//
	// If not called, the router will match the unencoded path to the routes.
	// For eg. "/path/foo%2Fbar/to" will match the path "/path/foo/bar/to"
	UseEncodedPath()

	// Group creates a new router group.
	Group(path string) RouterGroup

	// Any registers a route that matches all the HTTP methods.
	// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
	Any(path string, handler interface{}, r ...Renderer)
	// Get registers a new GET route with a matcher for the URL path of the get method.
	Get(path string, handler interface{}, r ...Renderer)
	// Head registers a new HEAD route with a matcher for the URL path of the get method.
	Head(path string, handler interface{}, r ...Renderer)
	// Post registers a new POST route with a matcher for the URL path of the get method.
	Post(path string, handler interface{}, r ...Renderer)
	// Put registers a new PUT route with a matcher for the URL path of the get method.
	Put(path string, handler interface{}, r ...Renderer)
	// Patch registers a new PATCH route with a matcher for the URL path of the get method.
	Patch(path string, handler interface{}, r ...Renderer)
	// Delete registers a new DELETE route with a matcher for the URL path of the get method.
	Delete(path string, handler interface{}, r ...Renderer)
	// Connect registers a new CONNECT route with a matcher for the URL path of the get method.
	Connect(path string, handler interface{}, r ...Renderer)
	// Options registers a new OPTIONS route with a matcher for the URL path of the get method.
	Options(path string, handler interface{}, r ...Renderer)
	// Trace registers a new TRACE route with a matcher for the URL path of the get method.
	Trace(path string, handler interface{}, r ...Renderer)
}

type routerGroup struct {
	basePath string
	router   *mux.Router
	renderer Renderer
}

// BasePath returns the base path of router group.
func (s *routerGroup) BasePath() string {
	return s.basePath
}

// NotFound to be used when no route matches.
// This can be used to render your own 404 Not Found errors.
func (s *routerGroup) NotFound(handler http.Handler) {
	s.router.NotFoundHandler = handler
}

// MethodNotAllowed to be used when the request method does not match the route.
// This can be used to render your own 405 Method Not Allowed errors.
func (s *routerGroup) MethodNotAllowed(handler http.Handler) {
	s.router.MethodNotAllowedHandler = handler
}

// Renderer to be used Response renderer in default.
func (s *routerGroup) Renderer(renderer Renderer) {
	s.renderer = renderer
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
func (s *routerGroup) StrictSlash(value bool) {
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
func (s *routerGroup) SkipClean(value bool) {
	s.router.SkipClean(value)
}

// UseEncodedPath tells the router to match the encoded original path
// to the routes.
// For eg. "/path/foo%2Fbar/to" will match the path "/path/{var}/to".
//
// If not called, the router will match the unencoded path to the routes.
// For eg. "/path/foo%2Fbar/to" will match the path "/path/foo/bar/to"
func (s *routerGroup) UseEncodedPath() {
	s.router.UseEncodedPath()
}

// ServeHTTP dispatches the handler registered in the matched route.
//
// When there is a match, the route variables can be retrieved calling
// Vars(request).
func (s *routerGroup) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// Use appends a MiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify requests and/or responses, and are executed in the order that they are applied to the Router.
func (s *routerGroup) Use(mwf ...MiddlewareFunc) {
	s.router.Use(mwf...)
}

// Group creates a new router group.
func (s *routerGroup) Group(path string) RouterGroup {
	group := &routerGroup{
		basePath: joinPaths(s.basePath, path),
		router:   s.router.PathPrefix(path).Subrouter(),
		renderer: s.renderer,
	}
	group.router.NotFoundHandler = s.router.NotFoundHandler
	group.router.MethodNotAllowedHandler = s.router.MethodNotAllowedHandler
	return group
}

// Bind registers a new route with a matcher for the URL path.
// Automatic binding request to handler input params, following functions:
//
// func(ctx context.Context)
//
// func(ctx context.Context) R
//
// func(ctx context.Context) error
//
// func(ctx context.Context, req T) R
//
// func(ctx context.Context, req T) error
//
// func(ctx context.Context, req T) (R, error)
func (s *routerGroup) Bind(path string, handler interface{}, r ...Renderer) *mux.Route {
	var renderer = s.renderer
	if len(r) > 0 {
		renderer = r[0]
	}
	return s.Handle(path, Bind(handler, renderer))
}

// Handle registers a new route with a matcher for the URL path.
// See Route.Path() and Route.Handler().
func (s *routerGroup) Handle(path string, handler http.Handler) *mux.Route {
	return s.router.Handle(path, handler)
}

// HandleFunc registers a new route with a matcher for the URL path.
// See Route.Path() and Route.HandlerFunc().
func (s *routerGroup) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.router.HandleFunc(path, f)
}

// Walk walks the router and all its sub-routers, calling walkFn for each route
// in the tree. The routes are walked in the order they were added. Sub-routers
// are explored depth-first.
func (s *routerGroup) Walk(walkFn mux.WalkFunc) error {
	return s.router.Walk(walkFn)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (s *routerGroup) Any(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...)
}

// Get registers a new GET route with a matcher for the URL path of the get method.
func (s *routerGroup) Get(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodGet)
}

// Head registers a new HEAD route with a matcher for the URL path of the get method.
func (s *routerGroup) Head(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodHead)
}

// Post registers a new POST route with a matcher for the URL path of the get method.
func (s *routerGroup) Post(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodPost)
}

// Put registers a new PUT route with a matcher for the URL path of the get method.
func (s *routerGroup) Put(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodPut)
}

// Patch registers a new PATCH route with a matcher for the URL path of the get method.
func (s *routerGroup) Patch(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodPatch)
}

// Delete registers a new DELETE route with a matcher for the URL path of the get method.
func (s *routerGroup) Delete(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodDelete)
}

// Connect registers a new CONNECT route with a matcher for the URL path of the get method.
func (s *routerGroup) Connect(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodConnect)
}

// Options registers a new OPTIONS route with a matcher for the URL path of the get method.
func (s *routerGroup) Options(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodOptions)
}

// Trace registers a new TRACE route with a matcher for the URL path of the get method.
func (s *routerGroup) Trace(path string, handler interface{}, r ...Renderer) {
	s.Bind(path, handler, r...).Methods(http.MethodTrace)
}
