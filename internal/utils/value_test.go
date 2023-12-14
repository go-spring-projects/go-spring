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

package utils

import (
	"reflect"
	"testing"

	"go-spring.dev/spring/internal/utils/assert"
	"go-spring.dev/spring/internal/utils/testdata"
)

func TestPatchValue(t *testing.T) {
	var r struct{ v int }
	v := reflect.ValueOf(&r)
	v = v.Elem().Field(0)
	assert.Panic(t, func() {
		v.SetInt(4)
	}, "using value obtained using unexported field")
	v = PatchValue(v)
	v.SetInt(4)
}

func TestIndirect(t *testing.T) {
	var r struct{ v int }
	typ := reflect.TypeOf(r)
	assert.Equal(t, Indirect(typ), reflect.TypeOf(r))
	typ = reflect.TypeOf(&r)
	assert.Equal(t, Indirect(typ), reflect.TypeOf(r))
}

//go:noinline
func fnNoArgs() {}

//go:noinline
func fnWithArgs(i int) {}

type receiver struct{}

//go:noinline
func (r receiver) fnNoArgs() {}

//go:noinline
func (r receiver) fnWithArgs(i int) {}

//go:noinline
func (r *receiver) ptrFnNoArgs() {}

//go:noinline
func (r *receiver) ptrFnWithArgs(i int) {}

func TestFileLine(t *testing.T) {
	testcases := []struct {
		fn     interface{}
		file   string
		line   int
		fnName string
	}{
		{
			fn:     fnNoArgs,
			file:   "utils/value_test.go",
			line:   47,
			fnName: "fnNoArgs",
		},
		{
			fnWithArgs,
			"utils/value_test.go",
			50,
			"fnWithArgs",
		},
		{
			receiver{}.fnNoArgs,
			"utils/value_test.go",
			52,
			"receiver.fnNoArgs",
		},
		{
			receiver.fnNoArgs,
			"utils/value_test.go",
			52,
			"receiver.fnNoArgs",
		},
		{
			receiver{}.fnWithArgs,
			"utils/value_test.go",
			54,
			"receiver.fnWithArgs",
		},
		{
			receiver.fnWithArgs,
			"utils/value_test.go",
			54,
			"receiver.fnWithArgs",
		},
		{
			(&receiver{}).ptrFnNoArgs,
			"utils/value_test.go",
			56,
			"(*receiver).ptrFnNoArgs",
		},
		{
			(*receiver).ptrFnNoArgs,
			"utils/value_test.go",
			56,
			"(*receiver).ptrFnNoArgs",
		},
		{
			(&receiver{}).ptrFnWithArgs,
			"utils/value_test.go",
			58,
			"(*receiver).ptrFnWithArgs",
		},
		{
			(*receiver).ptrFnWithArgs,
			"utils/value_test.go",
			58,
			"(*receiver).ptrFnWithArgs",
		},
		{
			testdata.FnNoArgs,
			"utils/testdata/pkg.go",
			19,
			"FnNoArgs",
		},
		{
			testdata.FnWithArgs,
			"utils/testdata/pkg.go",
			21,
			"FnWithArgs",
		},
		{
			testdata.Receiver{}.FnNoArgs,
			"utils/testdata/pkg.go",
			25,
			"Receiver.FnNoArgs",
		},
		{
			testdata.Receiver{}.FnWithArgs,
			"utils/testdata/pkg.go",
			27,
			"Receiver.FnWithArgs",
		},
		{
			(&testdata.Receiver{}).PtrFnNoArgs,
			"utils/testdata/pkg.go",
			29,
			"(*Receiver).PtrFnNoArgs",
		},
		{
			(&testdata.Receiver{}).PtrFnWithArgs,
			"utils/testdata/pkg.go",
			31,
			"(*Receiver).PtrFnWithArgs",
		},
		{
			testdata.Receiver.FnNoArgs,
			"utils/testdata/pkg.go",
			25,
			"Receiver.FnNoArgs",
		},
		{
			testdata.Receiver.FnWithArgs,
			"utils/testdata/pkg.go",
			27,
			"Receiver.FnWithArgs",
		},
		{
			(*testdata.Receiver).PtrFnNoArgs,
			"utils/testdata/pkg.go",
			29,
			"(*Receiver).PtrFnNoArgs",
		},
		{
			(*testdata.Receiver).PtrFnWithArgs,
			"utils/testdata/pkg.go",
			31,
			"(*Receiver).PtrFnWithArgs",
		},
	}
	for _, c := range testcases {
		file, line, fnName := FileLineFromPC(c.fn)
		t.Log(file, line, fnName)
		//assert.String(t, file).HasSuffix(c.file)
		//assert.Equal(t, line, c.line)
		assert.Equal(t, fnName, c.fnName)
	}
}
