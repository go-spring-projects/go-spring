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
	"testing"

	"github.com/go-spring-projects/go-spring/conf"
	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestBool(t *testing.T) {

	var u Bool
	assert.Equal(t, u.Value(), false)

	param := conf.BindParam{
		Key:  "bool",
		Path: "bool",
		Tag: conf.ParsedTag{
			Key: "bool",
		},
	}

	p := assert.Must(conf.Map(nil))
	err := u.OnRefresh(p, param)
	assert.Error(t, err, "bind bool error: property \"bool\": not exist")

	_ = p.Set("bool", true)
	err = u.OnRefresh(p, param)
	assert.Equal(t, u.Value(), true)

	b, err := json.Marshal(&u)
	assert.Nil(t, err)
	assert.Equal(t, string(b), "true")
}
