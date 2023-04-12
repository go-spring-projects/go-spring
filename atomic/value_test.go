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

	"github.com/limpo1989/go-spring/utils/assert"
)

func TestValue(t *testing.T) {

	var x Value
	assert.Equal(t, x.Load(), nil)

	s1 := &properties{Data: 1}
	x.Store(s1)
	assert.Equal(t, x.Load(), s1)

	x.SetMarshalJSON(func(i interface{}) ([]byte, error) {
		s := i.(*properties)
		return json.Marshal(s)
	})

	bytes, _ := json.Marshal(&x)
	assert.Equal(t, string(bytes), `{"Data":1}`)
}
