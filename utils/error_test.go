/*
 * Copyright 2012-2019 the original author or authors.
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
	"testing"

	"github.com/limpo1989/go-spring/utils/assert"
)

func TestError(t *testing.T) {

	e0 := Error(FileLine(), "error: 0")
	t.Log(e0)
	assert.Error(t, e0, ".*/error_test.go:27 error: 0")

	e1 := Errorf(FileLine(), "error: %d", 1)
	t.Log(e1)
	assert.Error(t, e1, ".*/error_test.go:31 error: 1")

	e2 := Wrap(e0, FileLine(), "error: 0")
	t.Log(e2)
	assert.Error(t, e2, ".*/error_test.go:35 error: 0; .*/error_test.go:27 error: 0")

	e3 := Wrapf(e1, FileLine(), "error: %d", 1)
	t.Log(e3)
	assert.Error(t, e3, ".*/error_test.go:39 error: 1; .*/error_test.go:31 error: 1")

	e4 := Wrap(e2, FileLine(), "error: 0")
	t.Log(e4)
	assert.Error(t, e4, ".*/error_test.go:43 error: 0; .*/error_test.go:35 error: 0; .*/error_test.go:27 error: 0")

	e5 := Wrapf(e3, FileLine(), "error: %d", 1)
	t.Log(e5)
	assert.Error(t, e5, ".*/error_test.go:47 error: 1; .*/error_test.go:39 error: 1; .*/error_test.go:31 error: 1")
}
