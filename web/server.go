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
	"net/http"
)

// A Server defines parameters for running an HTTP server.
type Server struct {
	options Options
	httpSvr *http.Server
	Router
}

// NewServer returns a new server instance.
func NewServer(options Options) *Server {

	var addr = options.Addr
	if 0 == len(addr) {
		addr = ":8080" // default port: 8080
	}

	var router = options.Router
	if nil == router {
		router = NewRouter()
	}

	svr := &Server{
		options: options,
		httpSvr: &http.Server{
			Addr:              addr,
			Handler:           router,
			TLSConfig:         options.TlsConfig(),
			ReadTimeout:       options.ReadTimeout,
			ReadHeaderTimeout: options.ReadHeaderTimeout,
			WriteTimeout:      options.WriteTimeout,
			IdleTimeout:       options.IdleTimeout,
			MaxHeaderBytes:    options.MaxHeaderBytes,
		},
		Router: router,
	}

	return svr
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
