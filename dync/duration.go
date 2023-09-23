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
	"time"

	"github.com/go-spring-projects/go-spring/conf"
)

var _ conf.Value = (*Duration)(nil)

// A Duration is an atomic time.Duration value that can be dynamic refreshed.
type Duration struct {
	v atomic.Int64
}

// Store atomically stores val.
func (x *Duration) Store(v time.Duration) {
	x.v.Store(int64(v))
}

// Value returns the stored time.Duration value.
func (x *Duration) Value() time.Duration {
	return time.Duration(x.v.Load())
}

// OnRefresh refreshes the stored value.
func (x *Duration) OnRefresh(p *conf.Properties, param conf.BindParam) error {
	var d time.Duration
	if err := p.Bind(&d, conf.Param(param)); err != nil {
		return err
	}
	x.v.Store(int64(d))
	return nil
}

// MarshalJSON returns the JSON encoding of x.
func (x *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
