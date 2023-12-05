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

	"github.com/gorilla/mux"
)

var (
	// ErrMethodMismatch is returned when the method in the request does not match
	// the method defined against the route.
	ErrMethodMismatch = mux.ErrMethodMismatch
	// ErrNotFound is returned when no route match is found.
	ErrNotFound = mux.ErrNotFound
	// SkipRouter is used as a return value from WalkFuncs to indicate that the
	// router that walk is about to descend down to should be skipped.
	SkipRouter = mux.SkipRouter
)

type (
	// Router registers routes to be matched and dispatches a handler.
	//
	// It implements the http.Handler interface, so it can be registered to serve
	// requests:
	//
	//	var router = mux.NewRouter()
	//
	//	func main() {
	//	    http.Handle("/", router)
	//	}
	//
	// Or, for Google App Engine, register it in a init() function:
	//
	//	func init() {
	//	    http.Handle("/", router)
	//	}
	//
	// This will send all incoming requests to the router.
	Router = mux.Router
	// Route stores information to match a request and build URLs.
	Route = mux.Route
	// RouteMatch stores information about a matched route.
	RouteMatch = mux.RouteMatch
	// BuildVarsFunc is the function signature used by custom build variable
	// functions (which can modify route variables before a route's URL is built).
	BuildVarsFunc = mux.BuildVarsFunc
	// MatcherFunc is the function signature used by custom matchers.
	MatcherFunc = mux.MatcherFunc
	// WalkFunc is the type of the function called for each route visited by Walk.
	// At every invocation, it is given the current route, and the current router,
	// and a list of ancestor routes that lead to the current route.
	WalkFunc = mux.WalkFunc
	// MiddlewareFunc is a function which receives an http.Handler and returns another http.Handler.
	// Typically, the returned handler is a closure which does something with the http.ResponseWriter and http.Request passed
	// to it, and then calls the handler passed as parameter to the MiddlewareFunc.
	MiddlewareFunc = mux.MiddlewareFunc
)

// NewRouter returns a new router instance.
func NewRouter() *Router {
	return mux.NewRouter()
}

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// CurrentRoute returns the matched route for the current request, if any.
// This only works when called inside the handler of the matched route
// because the matched route is stored in the request context which is cleared
// after the handler returns.
func CurrentRoute(r *http.Request) *Route {
	return mux.CurrentRoute(r)
}
