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

var _ conf.Value = (*Map[int, any])(nil)

type Map[K comparable, V any] struct {
	v atomic.Pointer[map[K]V]
}

// Store atomically stores val.
func (x *Map[K, V]) Store(v map[K]V) {
	x.v.Store(&v)
}

// Value returns the stored map[K]V value.
func (x *Map[K, V]) Value() map[K]V {
	if v := x.v.Load(); nil != v {
		return *v
	}
	return nil
}

// OnRefresh refreshes the stored value.
func (x *Map[K, V]) OnRefresh(p *conf.Properties, param conf.BindParam) error {
	var m map[K]V
	if err := p.Bind(&m, conf.Param(param)); err != nil {
		return err
	}
	x.v.Store(&m)
	return nil
}

func (x *Map[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
