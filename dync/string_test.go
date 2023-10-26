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

func TestString(t *testing.T) {

	var u String
	assert.Equal(t, u.Value(), "")

	param := conf.BindParam{
		Key:  "s",
		Path: "string",
		Tag: conf.ParsedTag{
			Key: "s",
		},
	}

	p := assert.Must(conf.Map(nil))
	err := u.OnRefresh(p, param)
	assert.Error(t, err, "bind string error: property \"s\": not exist")

	_ = p.Set("s", "abc")

	param.Validate = ""
	err = u.OnRefresh(p, param)
	assert.Equal(t, u.Value(), "abc")

	b, err := json.Marshal(&u)
	assert.Nil(t, err)
	assert.Equal(t, string(b), "\"abc\"")
}
