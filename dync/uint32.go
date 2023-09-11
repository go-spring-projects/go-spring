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

	"github.com/limpo1989/go-spring/conf"
)

var _ conf.Value = (*Uint32)(nil)

// An Uint32 is an atomic uint32 value that can be dynamic refreshed.
type Uint32 struct {
	v atomic.Uint32
}

// Store atomically stores val.
func (x *Uint32) Store(v uint32) {
	x.v.Store(v)
}

// Value returns the stored uint32 value.
func (x *Uint32) Value() uint32 {
	return x.v.Load()
}

// OnRefresh refreshes the stored value.
func (x *Uint32) OnRefresh(p *conf.Properties, param conf.BindParam) error {
	var u uint32
	if err := p.Bind(&u, conf.Param(param)); err != nil {
		return err
	}
	x.v.Store(u)
	return nil
}

// MarshalJSON returns the JSON encoding of x.
func (x *Uint32) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
