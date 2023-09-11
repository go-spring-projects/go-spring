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
	"math"
	"sync/atomic"

	"github.com/limpo1989/go-spring/conf"
)

var _ conf.Value = (*Float32)(nil)

// A Float32 is an atomic float32 value that can be dynamic refreshed.
type Float32 struct {
	v atomic.Uint32
}

// Store atomically stores val.
func (x *Float32) Store(v float32) {
	x.v.Store(math.Float32bits(v))
}

// Value returns the stored float32 value.
func (x *Float32) Value() float32 {
	return math.Float32frombits(x.v.Load())
}

// OnRefresh refreshes the stored value.
func (x *Float32) OnRefresh(p *conf.Properties, param conf.BindParam) error {
	var f float32
	if err := p.Bind(&f, conf.Param(param)); err != nil {
		return err
	}
	x.v.Store(math.Float32bits(f))
	return nil
}

// MarshalJSON returns the JSON encoding of x.
func (x *Float32) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
