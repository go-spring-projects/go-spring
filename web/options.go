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
	"crypto/tls"
	"time"
)

type Options struct {
	// Addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port". If empty, ":http" (port 8080) is used.
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	Addr string `json:"addr" value:"${addr:=}"`

	// CertFile containing a certificate and matching private key for the
	// server must be provided if neither the Server's
	// TLSConfig.Certificates nor TLSConfig.GetCertificate are populated.
	// If the certificate is signed by a certificate authority, the
	// certFile should be the concatenation of the server's certificate,
	// any intermediates, and the CA's certificate.
	CertFile string `json:"cert-file" value:"${cert-file:=}"`

	// KeyFile containing a private key file.
	KeyFile string `json:"key-file" value:"${key-file:=}"`

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration `json:"read-timeout" value:"${read-timeout:=0s}"`

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration `json:"read-header-timeout" value:"${read-header-timeout:=0s}"`

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout time.Duration `json:"write-timeout" value:"${write-timeout:=0s}"`

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	IdleTimeout time.Duration `json:"idle-timeout" value:"${idle-timeout:=0s}"`

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `json:"max-header-bytes" value:"${max-header-bytes:=0}"`

	// Router optionally specifies an external router.
	Router Router `json:"-"`
}

func (options Options) IsTls() bool {
	return len(options.CertFile) > 0 && len(options.KeyFile) > 0
}

func (options Options) TlsConfig() *tls.Config {
	if !options.IsTls() {
		return nil
	}

	return &tls.Config{
		GetCertificate: options.GetCertificate,
	}
}

func (options Options) GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(options.CertFile, options.KeyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}
