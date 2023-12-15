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

import "net/http"

// MiddlewareFunc is a function which receives an http.Handler and returns another http.Handler.
// Typically, the returned handler is a closure which does something with the http.ResponseWriter and http.Request passed
// to it, and then calls the handler passed as parameter to the MiddlewareFunc.
type MiddlewareFunc = func(next http.Handler) http.Handler

// Middlewares type is a slice of standard middleware handlers with methods
// to compose middleware chains and http.Handler's.
type Middlewares []MiddlewareFunc

// Handler builds and returns a http.Handler from the chain of middlewares,
// with `h http.Handler` as the final handler.
func (mws Middlewares) Handler(h http.Handler) http.Handler {
	return &chainHandler{Endpoint: h, chain: mws.chain(h), Middlewares: mws}
}

// HandlerFunc builds and returns a http.Handler from the chain of middlewares,
// with `h http.Handler` as the final handler.
func (mws Middlewares) HandlerFunc(h http.HandlerFunc) http.Handler {
	return &chainHandler{Endpoint: h, chain: mws.chain(h), Middlewares: mws}
}

// Build a http.Handler composed of an inline middlewares.
func (mws Middlewares) chain(handler http.Handler) http.Handler {
	if 0 == len(mws) {
		return handler
	}

	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

type chainHandler struct {
	Endpoint    http.Handler
	chain       http.Handler
	Middlewares Middlewares
}

func (c *chainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.chain.ServeHTTP(w, r)
}
