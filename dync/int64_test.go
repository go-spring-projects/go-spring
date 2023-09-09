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
	"testing"

	"github.com/limpo1989/go-spring/conf"
	"github.com/limpo1989/go-spring/utils/assert"
)

func TestInt64(t *testing.T) {

	var u Int64
	assert.Equal(t, u.Value(), int64(0))

	param := conf.BindParam{
		Key:  "int",
		Path: "int64",
		Tag: conf.ParsedTag{
			Key: "int",
		},
	}

	p := conf.Map(nil)
	err := u.OnRefresh(p, param)
	assert.Error(t, err, "bind int64 error: property \"int\": not exist")

	_ = p.Set("int", int64(3))

	param.Validate = `expr:"$>5"`
	err = u.OnRefresh(p, param)
	assert.Error(t, err, "validate failed on \"\\$\\>5\" for value 3")

	param.Validate = ""
	err = u.OnRefresh(p, param)
	assert.Equal(t, u.Value(), int64(3))

	b, err := json.Marshal(&u)
	assert.Nil(t, err)
	assert.Equal(t, string(b), "3")
}
