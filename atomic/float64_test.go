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

package atomic

import (
	"encoding/json"
	"testing"
	"unsafe"

	"github.com/limpo1989/go-spring/utils/assert"
)

func TestFloat64(t *testing.T) {

	// atomic.Float64 and uint64 occupy the same space
	assert.Equal(t, unsafe.Sizeof(Float64{}), uintptr(8))

	var f Float64
	assert.Equal(t, f.Load(), float64(0))

	v := f.Add(0.5)
	assert.Equal(t, v, 0.5)
	assert.Equal(t, f.Load(), 0.5)

	f.Store(1)
	assert.Equal(t, f.Load(), float64(1))

	old := f.Swap(2)
	assert.Equal(t, old, float64(1))
	assert.Equal(t, f.Load(), float64(2))

	swapped := f.CompareAndSwap(2, 3)
	assert.True(t, swapped)
	assert.Equal(t, f.Load(), float64(3))

	swapped = f.CompareAndSwap(2, 3)
	assert.False(t, swapped)
	assert.Equal(t, f.Load(), float64(3))

	bytes, _ := json.Marshal(&f)
	assert.Equal(t, string(bytes), "3")
}
