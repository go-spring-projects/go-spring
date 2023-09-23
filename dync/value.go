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

package dync

import (
	"encoding/json"
	"sync/atomic"

	"github.com/go-spring-projects/go-spring/conf"
)

var _ conf.Value = (*Value[any])(nil)

type Value[T any] struct {
	v atomic.Pointer[T]
}

// Store atomically stores val.
func (x *Value[T]) Store(v *T) {
	x.v.Store(v)
}

// Value returns the stored value.
func (x *Value[T]) Value() *T {
	return x.v.Load()
}

// OnRefresh refreshes the stored value.
func (x *Value[T]) OnRefresh(p *conf.Properties, param conf.BindParam) error {
	var v T
	if err := p.Bind(&v, conf.Param(param)); nil != err {
		return err
	}
	x.v.Store(&v)
	return nil
}

func (x *Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
