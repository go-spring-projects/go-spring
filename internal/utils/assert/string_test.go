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

package assert

import (
	"testing"
)

func TestString_EqualFold(t *testing.T) {
	runCase(t, func(g *MockT) {
		String(g, "hello, world!").EqualFold("Hello, World!")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't equal fold to 'xxx'"})
		String(g, "hello, world!").EqualFold("xxx")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't equal fold to 'xxx'; param (index=0)"})
		String(g, "hello, world!").EqualFold("xxx", "param (index=0)")
	})
}

func TestString_HasPrefix(t *testing.T) {
	runCase(t, func(g *MockT) {
		String(g, "hello, world!").HasPrefix("hello")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't have prefix 'xxx'"})
		String(g, "hello, world!").HasPrefix("xxx")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't have prefix 'xxx'; param (index=0)"})
		String(g, "hello, world!").HasPrefix("xxx", "param (index=0)")
	})
}

func TestString_HasSuffix(t *testing.T) {
	runCase(t, func(g *MockT) {
		String(g, "hello, world!").HasSuffix("world!")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't have suffix 'xxx'"})
		String(g, "hello, world!").HasSuffix("xxx")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't have suffix 'xxx'; param (index=0)"})
		String(g, "hello, world!").HasSuffix("xxx", "param (index=0)")
	})
}

func TestString_Contains(t *testing.T) {
	runCase(t, func(g *MockT) {
		String(g, "hello, world!").Contains("hello")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't contain substr 'xxx'"})
		String(g, "hello, world!").Contains("xxx")
	})
	runCase(t, func(g *MockT) {
		g.EXPECT().Error([]interface{}{"'hello, world!' doesn't contain substr 'xxx'; param (index=0)"})
		String(g, "hello, world!").Contains("xxx", "param (index=0)")
	})
}
