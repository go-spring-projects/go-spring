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
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-spring-projects/go-spring/internal/utils"
	"github.com/go-spring-projects/go-spring/web/binding"
)

type Renderer interface {
	Render(ctx *Context, err error, result interface{})
}

type RendererFunc func(ctx *Context, err error, result interface{})

func (fn RendererFunc) Render(ctx *Context, err error, result interface{}) {
	fn(ctx, err, result)
}

// Bind convert fn to HandlerFunc.
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
//
// func(writer http.ResponseWriter, request *http.Request)
func Bind(fn interface{}, render Renderer) http.HandlerFunc {

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	switch h := fn.(type) {
	case http.HandlerFunc:
		return warpHandlerCtx(h)
	case http.Handler:
		return warpHandlerCtx(h.ServeHTTP)
	case func(http.ResponseWriter, *http.Request):
		return warpHandlerCtx(h)
	default:
		// valid func
		if err := validMappingFunc(fnType); nil != err {
			panic(err)
		}
	}

	firstOutIsErrorType := 1 == fnType.NumOut() && utils.IsErrorType(fnType.Out(0))

	return func(writer http.ResponseWriter, request *http.Request) {

		// param of context
		webCtx := &Context{Writer: writer, Request: request}
		ctx := WithContext(request.Context(), webCtx)
		ctxValue := reflect.ValueOf(ctx)

		defer func() {
			if nil != request.MultipartForm {
				_ = request.MultipartForm.RemoveAll()
			}
			_ = request.Body.Close()
		}()

		var returnValues []reflect.Value
		var err error

		defer func() {
			if r := recover(); nil != r {
				if e, ok := r.(error); ok {
					err = fmt.Errorf("%s: %w", request.URL.Path, e)
				} else {
					err = fmt.Errorf("%s: %v", request.URL.Path, r)
				}

				// render error response
				render.Render(webCtx, err, nil)
			}
		}()

		switch fnType.NumIn() {
		case 1:
			returnValues = fnValue.Call([]reflect.Value{ctxValue})
		case 2:
			paramType := fnType.In(1)
			pointer := false
			if reflect.Ptr == paramType.Kind() {
				paramType = paramType.Elem()
				pointer = true
			}

			// new param instance with paramType.
			paramValue := reflect.New(paramType)
			// bind paramValue with request
			if err = binding.Bind(paramValue.Interface(), webCtx); nil != err {
				break
			}
			if !pointer {
				paramValue = paramValue.Elem()
			}
			returnValues = fnValue.Call([]reflect.Value{ctxValue, paramValue})
		default:
			panic("unreachable here")
		}

		var result interface{}

		if nil == err {
			switch len(returnValues) {
			case 0:
				// nothing
				return
			case 1:
				if firstOutIsErrorType {
					err, _ = returnValues[0].Interface().(error)
				} else {
					result = returnValues[0].Interface()
				}
			case 2:
				// check error
				result = returnValues[0].Interface()
				err, _ = returnValues[1].Interface().(error)
			default:
				panic("unreachable here")
			}
		}

		// render response
		render.Render(webCtx, err, result)
	}
}

func validMappingFunc(fnType reflect.Type) error {
	// func(ctx context.Context)
	// func(ctx context.Context) R
	// func(ctx context.Context) error
	// func(ctx context.Context, req T) R
	// func(ctx context.Context, req T) error
	// func(ctx context.Context, req T) (R, error)
	if !utils.IsFuncType(fnType) {
		return fmt.Errorf("%s: not a func", fnType.String())
	}

	if fnType.NumIn() < 1 || fnType.NumIn() > 2 {
		return fmt.Errorf("%s: invalid input parameter count", fnType.String())
	}

	if fnType.NumOut() > 2 {
		return fmt.Errorf("%s: invalid output parameter count", fnType.String())
	}

	if !utils.IsContextType(fnType.In(0)) {
		return fmt.Errorf("%s: first input param type (%s) must be context", fnType.String(), fnType.In(0).String())
	}

	if fnType.NumIn() > 1 {
		argType := fnType.In(1)
		if !(reflect.Struct == argType.Kind() || (reflect.Ptr == argType.Kind() && reflect.Struct == argType.Elem().Kind())) {
			return fmt.Errorf("%s: second input param type (%s) must be struct/*struct", fnType.String(), argType.String())
		}
	}

	switch fnType.NumOut() {
	case 0: // nothing
	case 1: // R | error
	case 2: // (R, error)
		if utils.IsErrorType(fnType.Out(0)) {
			return fmt.Errorf("%s: first output param type not be error", fnType.String())
		}

		if !utils.IsErrorType(fnType.Out(1)) {
			return fmt.Errorf("%s: second output type (%s) must a error", fnType.String(), fnType.Out(1).String())
		}
	}

	return nil
}

func warpHandlerCtx(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		webCtx := &Context{Writer: writer, Request: request}
		handler.ServeHTTP(writer, requestWithCtx(request, webCtx))
	}
}

func requestWithCtx(r *http.Request, webCtx *Context) *http.Request {
	ctx := WithContext(r.Context(), webCtx)
	return r.WithContext(ctx)
}
