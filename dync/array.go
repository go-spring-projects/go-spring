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

package dync

import (
	"encoding/json"
	"sync/atomic"

	"github.com/go-spring-projects/go-spring/conf"
)

var _ conf.Value = (*Array[any])(nil)

type Array[T any] struct {
	v atomic.Pointer[[]T]
}

// Store atomically stores val.
func (x *Array[T]) Store(v []T) {
	x.v.Store(&v)
}

// Value returns the stored []T value.
func (x *Array[T]) Value() []T {
	if v := x.v.Load(); nil != v {
		return *v
	}
	return []T{}
}

// OnRefresh refreshes the stored value.
func (x *Array[T]) OnRefresh(p *conf.Properties, param conf.BindParam) error {
	var array []T
	if err := p.Bind(&array, conf.Param(param)); err != nil {
		return err
	}
	x.v.Store(&array)
	return nil
}

func (x *Array[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
