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
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/limpo1989/go-spring/utils/assert"
)

func TestInt64(t *testing.T) {

	// atomic.Int64 and int64 occupy the same space
	assert.Equal(t, unsafe.Sizeof(Int64{}), uintptr(8))

	var i Int64
	assert.Equal(t, i.Load(), int64(0))

	v := i.Add(5)
	assert.Equal(t, v, int64(5))
	assert.Equal(t, i.Load(), int64(5))

	i.Store(1)
	assert.Equal(t, i.Load(), int64(1))

	old := i.Swap(2)
	assert.Equal(t, old, int64(1))
	assert.Equal(t, i.Load(), int64(2))

	swapped := i.CompareAndSwap(2, 3)
	assert.True(t, swapped)
	assert.Equal(t, i.Load(), int64(3))

	swapped = i.CompareAndSwap(2, 3)
	assert.False(t, swapped)
	assert.Equal(t, i.Load(), int64(3))

	bytes, _ := json.Marshal(&i)
	assert.Equal(t, string(bytes), "3")
}

func TestReflectInt64(t *testing.T) {

	var s struct {
		I Int64
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		addr := reflect.ValueOf(&s).Elem().Field(0).Addr()
		v, ok := addr.Interface().(*Int64)
		assert.True(t, ok)
		for i := 0; i < 10; i++ {
			v.Add(1)
		}
	}()
	go func() {
		defer wg.Done()
		addr := reflect.ValueOf(&s).Elem().Field(0).Addr()
		v, ok := addr.Interface().(*Int64)
		assert.True(t, ok)
		for i := 0; i < 10; i++ {
			v.Add(2)
		}
	}()
	wg.Wait()
	assert.Equal(t, int64(30), s.I.Load())
}
