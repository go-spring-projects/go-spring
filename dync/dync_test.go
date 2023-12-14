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
	"reflect"
	"testing"

	"go-spring.dev/spring/conf"
	"go-spring.dev/spring/internal/utils/assert"
)

func TestProperties(t *testing.T) {
	p := New()
	m := map[string]interface{}{
		"a": 123,
		"b": map[string]interface{}{
			"c": []int{1, 2, 3},
		},
	}
	err := p.Refresh(assert.Must(conf.Map(m)))
	assert.Nil(t, err)
	assert.Equal(t, p.Keys(), []string{"a", "b.c[0]", "b.c[1]", "b.c[2]"})
	assert.False(t, p.Has("c"))
	assert.True(t, p.Has("b.c"))
	assert.Equal(t, p.Get("a"), "123")
	assert.Equal(t, p.Get("d", conf.Def("456")), "456")
	s, err := p.Resolve("a=${a}")
	assert.Nil(t, err)
	assert.Equal(t, s, "a=123")
	var b struct {
		C []int `value:"${c}"`
	}
	err = p.Bind(&b, conf.Key("b"))
	assert.Nil(t, err)
	assert.Equal(t, b.C, []int{1, 2, 3})
}

type Config struct {
	Int   Int64               `value:"${int:=3}" expr:"$<6"`
	Float Float64             `value:"${float:=1.2}"`
	Map   Map[string, string] `value:"${map:=}"`
	Slice Array[string]       `value:"${slice:=}"`
}

func newTest() (*Properties, *Config, error) {
	p := New()
	cfg := new(Config)
	cfg.Slice.Store(make([]string, 0))
	cfg.Map.Store(make(map[string]string))
	err := p.BindValue(reflect.ValueOf(cfg), conf.BindParam{})
	if err != nil {
		return nil, nil, err
	}
	return p, cfg, nil
}

func TestDynamic(t *testing.T) {

	t.Run("default & success", func(t *testing.T) {
		p, cfg, err := newTest()
		assert.Nil(t, err)
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":3,"Float":1.2,"Map":{},"Slice":[]}`)

		err = p.Refresh(assert.Must(conf.Map(map[string]interface{}{
			"int":   1,
			"float": 5.4,
		})))
		assert.Nil(t, err)
		b, _ = json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":1,"Float":5.4,"Map":{},"Slice":[]}`)

		err = p.Refresh(assert.Must(conf.Map(map[string]interface{}{
			"int": 2,
		})))
		assert.Nil(t, err)
		b, _ = json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":2,"Float":1.2,"Map":{},"Slice":[]}`)
	})

	t.Run("validate error", func(t *testing.T) {
		p, _, _ := newTest()
		err := p.Refresh(assert.Must(conf.Map(map[string]interface{}{
			"int": 9,
		})))
		assert.Error(t, err, "validate failed on \"\\$<6\" for value 9")
		err = p.Refresh(assert.Must(conf.Map(map[string]interface{}{
			"int": "abc.123",
		})))
		assert.Error(t, err, "parsing \\\"abc.123\\\": invalid syntax")
	})

	t.Run("bind value error", func(t *testing.T) {
		p := New()
		err := p.Refresh(assert.Must(conf.Map(map[string]interface{}{
			"int": "abc.123",
		})))
		assert.Nil(t, err)

		cfg := new(Config)
		cfg.Slice.Store(make([]string, 0))
		cfg.Map.Store(make(map[string]string))
		err = p.BindValue(reflect.ValueOf(cfg), conf.BindParam{})
		assert.Error(t, err, "parsing \\\"abc.123\\\": invalid syntax")

		var param conf.BindParam
		_ = param.BindTag("${int}", "")
		err = p.BindValue(reflect.ValueOf(new(Int64)), param)
		assert.Error(t, err, "parsing \\\"abc.123\\\": invalid syntax")

		err = p.Refresh(assert.Must(conf.Map(map[string]interface{}{
			"int": 123,
		})))
		assert.Nil(t, err)
		err = p.BindValue(reflect.ValueOf(new(Int64)), param)
		assert.Nil(t, err)
	})

	t.Run("set value error", func(t *testing.T) {
		p, cfg, err := newTest()
		assert.Nil(t, err)
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":3,"Float":1.2,"Map":{},"Slice":[]}`)

		err = p.Set("map.name", "jok")
		assert.Nil(t, err)

		name, exists := cfg.Map.Value()["name"]
		assert.True(t, exists)
		assert.Equal(t, name, "jok")

		err = p.Set("map.name[", "jok")
		assert.Error(t, err, `invalid key \'map\.name\[\'`)

		err = p.Remove("map.name")
		assert.Nil(t, err)

		_, exists = cfg.Map.Value()["name"]
		assert.False(t, exists)
	})

}
