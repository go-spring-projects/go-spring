/*
 * Copyright 2019 the original author or authors.
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

// Package binding ...
package binding

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-spring-projects/go-spring/conf"
)

var ErrBinding = errors.New("binding failed")
var ErrValidate = errors.New("validate failed")

const (
	MIMEApplicationJSON = "application/json"
	MIMEApplicationXML  = "application/xml"
	MIMETextXML         = "text/xml"
	MIMEApplicationForm = "application/x-www-form-urlencoded"
	MIMEMultipartForm   = "multipart/form-data"
)

type Request interface {
	ContentType() string
	Header(key string) (string, bool)
	Cookie(name string) (string, bool)
	PathParam(name string) (string, bool)
	QueryParam(name string) (string, bool)
	FormParams() (url.Values, error)
	MultipartParams(maxMemory int64) (*multipart.Form, error)
	RequestBody() io.Reader
}

type BindScope int

const (
	BindScopeURI BindScope = iota
	BindScopeQuery
	BindScopeHeader
	BindScopeCookie
	BindScopeBody
)

var scopeTags = map[BindScope]string{
	BindScopeURI:    "path",
	BindScopeQuery:  "query",
	BindScopeHeader: "header",
	BindScopeCookie: "cookie",
}

var scopeGetters = map[BindScope]func(r Request, name string) (string, bool){
	BindScopeURI:    Request.PathParam,
	BindScopeQuery:  Request.QueryParam,
	BindScopeHeader: Request.Header,
	BindScopeCookie: Request.Cookie,
}

type BodyBinder func(i interface{}, r Request) error

var bodyBinders = map[string]BodyBinder{
	MIMEApplicationForm: BindForm,
	MIMEMultipartForm:   BindMultipartForm,
	MIMEApplicationJSON: BindJSON,
	MIMEApplicationXML:  BindXML,
	MIMETextXML:         BindXML,
}

func RegisterScopeTag(scope BindScope, tag string) {
	scopeTags[scope] = tag
}

func RegisterBodyBinder(mime string, binder BodyBinder) {
	bodyBinders[mime] = binder
}

// Bind checks the Method and Content-Type to select a binding engine automatically,
// Depending on the "Content-Type" header different bindings are used, for example:
//
//	"application/json" --> JSON binding
//	"application/xml"  --> XML binding
func Bind(i interface{}, r Request) error {
	if err := bindScope(i, r); err != nil {
		return fmt.Errorf("%w: %v", ErrBinding, err)
	}

	if err := bindBody(i, r); err != nil {
		return fmt.Errorf("%w: %v", ErrBinding, err)
	}

	if err := conf.ValidateStruct(i); nil != err {
		return fmt.Errorf("%w: %v", ErrValidate, err)
	}
	return nil
}

func bindBody(i interface{}, r Request) error {
	mediaType, _, err := mime.ParseMediaType(r.ContentType())
	if nil != err && !strings.Contains(err.Error(), "mime: no media type") {
		return err
	}
	binder, ok := bodyBinders[mediaType]
	if !ok {
		binder = bodyBinders[MIMEApplicationForm]
	}
	return binder(i, r)
}

func bindScope(i interface{}, r Request) error {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("%s: is not pointer", t.String())
	}

	et := t.Elem()
	if et.Kind() != reflect.Struct {
		return fmt.Errorf("%s: is not a struct pointer", t.String())
	}

	ev := reflect.ValueOf(i).Elem()
	for j := 0; j < ev.NumField(); j++ {
		fv := ev.Field(j)
		ft := et.Field(j)
		for scope := BindScopeURI; scope < BindScopeBody; scope++ {
			if err := bindScopeField(scope, fv, ft, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func bindScopeField(scope BindScope, v reflect.Value, field reflect.StructField, r Request) error {
	if tag, loaded := scopeTags[scope]; loaded {
		if name, ok := field.Tag.Lookup(tag); ok && name != "-" {
			if val, exists := scopeGetters[scope](r, name); exists {
				if err := bindData(v, val); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func bindData(v reflect.Value, val string) error {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 0, 0)
		if err != nil {
			return err
		}
		v.SetUint(u)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 0, 0)
		if err != nil {
			return err
		}
		v.SetInt(i)
		return nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		v.SetBool(b)
		return nil
	case reflect.String:
		v.SetString(val)
		return nil
	default:
		return fmt.Errorf("unsupported binding type %q", v.Type().String())
	}
}
