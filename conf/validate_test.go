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

package conf

import (
	"testing"

	"go-spring.dev/spring/internal/utils/assert"
)

var empty = &emptyValidator{}

func init() {
	Register("empty", empty)
}

type emptyValidator struct {
	count int
}

func (d *emptyValidator) reset() {
	d.count = 0
}

func (d *emptyValidator) Field(tag string, i interface{}) error {
	d.count++
	return nil
}

func TestValidateField(t *testing.T) {
	i := 6

	err := Validate("empty:\"\"", i)
	assert.Nil(t, err)
	assert.Equal(t, empty.count, 1)

	err = Validate("expr:\"$>=3\"", i)
	assert.Nil(t, err)

	err = Validate("expr:\"$<3\"", i)
	assert.Error(t, err, "validate failed on \"\\$<3\" for value 6")

	err = Validate("expr:\"$<3\"", "abc")
	assert.Error(t, err, "invalid operation\\: string \\< int \\(1:2\\)")
}

func TestValidateStruct(t *testing.T) {
	type testForm struct {
		Age     int `expr:"$>=18"`
		Summary struct {
			Weight int `expr:"$>100"`
		}
		Skip       *struct{}
		unexported struct{}
	}

	tf1 := testForm{Age: 18}
	tf1.Summary.Weight = 101
	err := ValidateStruct(tf1)
	assert.Nil(t, err)

	tf2 := testForm{Age: 17}
	err = ValidateStruct(tf2)
	assert.Error(t, err, "validate failed on \"\\$>=18\" for value 17")

	tf3 := testForm{Age: 18}
	err = ValidateStruct(tf3)
	assert.Error(t, err, "validate failed on \"\\$>100\" for value 0")

}
