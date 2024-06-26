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
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"go-spring.dev/spring/internal/utils/assert"
)

type Point struct {
	X, Y int
}

func init() {
	RegisterConverter(func(val string) (Point, error) {
		if !(strings.HasPrefix(val, "(") && strings.HasSuffix(val, ")")) {
			return Point{}, errors.New("illegal format")
		}
		ss := strings.Split(val[1:len(val)-1], ",")
		x, err := strconv.Atoi(ss[0])
		if err != nil {
			return Point{}, err
		}
		y, err := strconv.Atoi(ss[1])
		if err != nil {
			return Point{}, err
		}
		return Point{X: x, Y: y}, nil
	})
}

func TestParseTag(t *testing.T) {
	var testcases = []struct {
		Tag   string
		Error string
		Data  string
	}{
		{
			Tag:   "||",
			Error: `parse tag '\|\|' error: invalid syntax`,
		},
		{
			Tag:   "a||",
			Error: `parse tag 'a\|\|' error: invalid syntax`,
		},
		{
			Tag:   "a}||",
			Error: `parse tag 'a}\|\|' error: invalid syntax`,
		},
		{
			Tag:  "${}||",
			Data: "${}",
		},
		{
			Tag:  "${}||k",
			Data: "${}||k",
		},
		{
			Tag:  "${a}||",
			Data: "${a}",
		},
		{
			Tag:  "${a}||k",
			Data: "${a}||k",
		},
		{
			Tag:  "${a:=}||",
			Data: "${a:=}",
		},
		{
			Tag:  "${a:=}||k",
			Data: "${a:=}||k",
		},
		{
			Tag:  "${a:=b}||",
			Data: "${a:=b}",
		},
		{
			Tag:  "${a:=b}||k",
			Data: "${a:=b}||k",
		},
	}
	for _, c := range testcases {
		tag, err := ParseTag(c.Tag)
		if c.Error != "" {
			assert.Error(t, err, c.Error)
			continue
		}
		assert.Nil(t, err)
		assert.Equal(t, tag.String(), c.Data)
		s, err := ParseTag(tag.String())
		assert.Nil(t, err)
		assert.Equal(t, s, tag)
	}
}

func TestBindTag(t *testing.T) {

	param := BindParam{}
	err := param.BindTag("{}", "")
	assert.Error(t, err, "parse tag '\\{\\}' error: invalid syntax")

	param = BindParam{}
	err = param.BindTag("${}", "")
	assert.Nil(t, err)
	assert.Equal(t, param, BindParam{
		Key: "",
		Tag: ParsedTag{
			Key: "",
		},
	})

	param = BindParam{
		Key: "s",
	}
	err = param.BindTag("${}", "")
	assert.Nil(t, err)
	assert.Equal(t, param, BindParam{
		Key: "s",
		Tag: ParsedTag{
			Key: "",
		},
	})

	param = BindParam{}
	err = param.BindTag("${}", "")
	assert.Nil(t, err)
	assert.Equal(t, param, BindParam{})

	param = BindParam{}
	err = param.BindTag("${a:=b}", "")
	assert.Nil(t, err)
	assert.Equal(t, param, BindParam{
		Key: "a",
		Tag: ParsedTag{
			Key:    "a",
			Def:    "b",
			HasDef: true,
		},
	})

	param = BindParam{
		Key: "s",
	}
	err = param.BindTag("${a:=b}", "")
	assert.Nil(t, err)
	assert.Equal(t, param, BindParam{
		Key: "s.a",
		Tag: ParsedTag{
			Key:    "a",
			Def:    "b",
			HasDef: true,
		},
	})
}

type PtrStruct struct {
	Int int `value:"${int}"`
}

type CommonStruct struct {
	Int           int           `value:"${int}"`
	Ints          []int         `value:"${ints}"`
	Uint          uint          `value:"${uint:=3}"`
	Uints         []uint        `value:"${uints:=1,2,3}"`
	Float         float64       `value:"${float:=3}"`
	Floats        []float64     `value:"${floats:=1,2,3}"`
	Bool          bool          `value:"${bool:=true}"`
	Bools         []bool        `value:"${bools:=true,false}"`
	String        string        `value:"${string:=abc}"`
	Strings       []string      `value:"${strings:=abc,def,ghi}"`
	Time          time.Time     `value:"${time:=2017-06-17 13:20:15 UTC}"`
	Duration      time.Duration `value:"${duration:=5s}"`
	dontWired     string
	dontWiredTime time.Time `value:"-"`
}

type FightExpansion struct {
	Round     int32   `value:"${round}" json:"round" validate:"min=1"`
	StartTime string  `value:"${start-time}" json:"start_time"`
	StopTime  string  `value:"${stop-time}" json:"stop_time"`
	Min       float64 `value:"${min}" json:"min" validate:"nonzero"`
	Max       float64 `value:"${max}" json:"max" validate:"nonzero"`
}

type NestedStruct struct {
	*PtrStruct
	CommonStruct
	Struct          CommonStruct
	Nested          CommonStruct     `value:"${nested}"`
	Int             int              `value:"${int}"`
	Ints            []int            `value:"${ints}"`
	Uint            uint             `value:"${uint:=3}"`
	Uints           []uint           `value:"${uints:=1,2,3}"`
	Float           float64          `value:"${float:=3}"`
	Floats          []float64        `value:"${floats:=1,2,3}"`
	Bool            bool             `value:"${bool:=true}"`
	Bools           []bool           `value:"${bools:=true,false}"`
	String          string           `value:"${string:=abc}"`
	Strings         []string         `value:"${strings:=abc,def,ghi}"`
	Time            time.Time        `value:"${time:=2017-06-17 13:20:15 UTC}"`
	Duration        time.Duration    `value:"${duration:=5s}"`
	FightExpansions []FightExpansion `value:"${fight-expansions:=}"`
}

func TestBind_InvalidValue(t *testing.T) {

	t.Run("invalid", func(t *testing.T) {
		var f float64
		err := assert.Must(Map(nil)).Bind(&f, Tag("a:=b"))
		assert.Error(t, err, "parse tag 'a:=b' error: invalid syntax")
	})

	t.Run("int", func(t *testing.T) {
		var i int
		err := assert.Must(Map(nil)).Bind(i)
		assert.Error(t, err, "i should be a ptr")
	})

	t.Run("chan", func(t *testing.T) {
		c := make(chan int)
		key := Key("chan")
		err := assert.Must(Map(nil)).Bind(&c, key)
		assert.Error(t, err, "bind chan int error: .*")
	})

	t.Run("array", func(t *testing.T) {
		var s [3]int
		key := Key("array")
		err := assert.Must(Map(nil)).Bind(&s, key)
		assert.Error(t, err, "bind \\[3\\]int error: use slice instead of array")
	})

	t.Run("complex", func(t *testing.T) {
		var c complex64
		tag := Tag("${complex:=i+3}")
		err := assert.Must(Map(nil)).Bind(&c, tag)
		assert.Error(t, err, "bind complex64 error: unsupported bind type \"complex64\"")
	})

	t.Run("pointer", func(t *testing.T) {
		var s struct {
			PtrStruct *PtrStruct `value:"${ptr}"`
		}
		err := assert.Must(Map(nil)).Bind(&s)
		assert.Error(t, err, "bind .* error: target should be value type")
	})
}

func TestBind_BindParam(t *testing.T) {
	p := assert.Must(Map(map[string]interface{}{
		"i": 3,
	}))
	param := BindParam{
		Key: "i",
		Tag: ParsedTag{
			Key: "i",
		},
	}
	var i int
	err := p.Bind(&i, Param(param))
	assert.Nil(t, err)
	assert.Equal(t, i, 3)
}

func TestBind_SingleValue(t *testing.T) {

	t.Run("uint", func(t *testing.T) {
		var u uint

		key := Key("uint")
		err := assert.Must(Map(nil)).Bind(&u, key)
		assert.Error(t, err, "bind uint error: property \"uint\": not exist")

		tag := Tag("${uint:=3}")
		err = assert.Must(Map(nil)).Bind(&u, tag)
		assert.Nil(t, err)
		assert.Equal(t, u, uint(3))

		err = assert.Must(Map(map[string]interface{}{
			"uint": 5,
		})).Bind(&u, tag)
		assert.Nil(t, err)
		assert.Equal(t, u, uint(5))

		err = assert.Must(Map(map[string]interface{}{
			"uint": "abc",
		})).Bind(&u, tag)
		assert.Error(t, err, "bind uint error: strconv.ParseUint: parsing \"abc\": invalid syntax")
	})

	t.Run("int", func(t *testing.T) {
		var i int

		key := Key("int")
		err := assert.Must(Map(nil)).Bind(&i, key)
		assert.Error(t, err, "bind int error: property \"int\": not exist")

		tag := Tag("${int:=3}")
		err = assert.Must(Map(nil)).Bind(&i, tag)
		assert.Nil(t, err)
		assert.Equal(t, i, 3)

		err = assert.Must(Map(map[string]interface{}{
			"int": 5,
		})).Bind(&i, tag)
		assert.Nil(t, err)
		assert.Equal(t, i, 5)

		err = assert.Must(Map(map[string]interface{}{
			"int": "abc",
		})).Bind(&i, tag)
		assert.Error(t, err, "bind int error: strconv.ParseInt: parsing \"abc\": invalid syntax")
	})

	t.Run("float", func(t *testing.T) {
		var f float32

		key := Key("float")
		err := assert.Must(Map(nil)).Bind(&f, key)
		assert.Error(t, err, "bind float32 error: property \"float\": not exist")

		tag := Tag("${float:=3}")
		err = assert.Must(Map(nil)).Bind(&f, tag)
		assert.Nil(t, err)
		assert.Equal(t, f, float32(3))

		err = assert.Must(Map(map[string]interface{}{
			"float": 5,
		})).Bind(&f, tag)
		assert.Nil(t, err)
		assert.Equal(t, f, float32(5))

		err = assert.Must(Map(map[string]interface{}{
			"float": "abc",
		})).Bind(&f, tag)
		assert.Error(t, err, "bind float32 error: strconv.ParseFloat: parsing \"abc\": invalid syntax")
	})

	t.Run("bool", func(t *testing.T) {
		var b bool

		key := Key("bool")
		err := assert.Must(Map(nil)).Bind(&b, key)
		assert.Error(t, err, "bind bool error: property \"bool\": not exist")

		tag := Tag("${bool:=false}")
		err = assert.Must(Map(nil)).Bind(&b, tag)
		assert.Nil(t, err)
		assert.Equal(t, b, false)

		err = assert.Must(Map(map[string]interface{}{
			"bool": true,
		})).Bind(&b, tag)
		assert.Nil(t, err)
		assert.Equal(t, b, true)

		err = assert.Must(Map(map[string]interface{}{
			"bool": "abc",
		})).Bind(&b, tag)
		assert.Error(t, err, "bind bool error: strconv.ParseBool: parsing \"abc\": invalid syntax")
	})

	t.Run("string", func(t *testing.T) {
		var s string

		tag := Tag("${string:=abc}")
		err := assert.Must(Map(nil)).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, "abc")

		err = assert.Must(Map(map[string]interface{}{
			"string": "def",
		})).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, "def")
	})

	t.Run("struct", func(t *testing.T) {
		var s NestedStruct

		tag := Tag("${struct:=abc,123}")
		err := assert.Must(Map(nil)).Bind(&s, tag)
		assert.Error(t, err, "bind .* error: struct can't have a non empty default value")

		tag = Tag("${struct:=}")
		err = assert.Must(Map(nil)).Bind(&s, tag)
		assert.Error(t, err, "bind NestedStruct error: bind NestedStruct.CommonStruct error: bind NestedStruct.CommonStruct.Int error: property \"struct.int\": not exist")

		tag = Tag("${struct:=}")
		err = assert.Must(Map(map[string]interface{}{
			"struct": map[string]interface{}{
				"int":  1,
				"ints": []int{1, 2, 3},
				"nested": map[string]interface{}{
					"int":           1,
					"ints":          "1,2,3",
					"dontWired":     "123456789",
					"dontWiredTime": "Mon May  6 23:31:43 EDT 2024",
				},
				"dontWired":     "987654321",
				"dontWiredTime": "Mon May  6 23:31:43 EDT 2024",
			},
		})).Bind(&s, tag)
		assert.Error(t, err, "bind NestedStruct error: bind NestedStruct.Struct error: bind NestedStruct.Struct.Int error: property \"struct.Struct.int\": not exist")
		assert.Equal(t, s.dontWired, "")
		assert.Equal(t, s.dontWiredTime, time.Time{})
		assert.Equal(t, s.Struct.dontWired, "")
		assert.Equal(t, s.Struct.dontWiredTime, time.Time{})

		m := map[string]interface{}{
			"struct": map[string]interface{}{
				"int":  1,
				"ints": []int{1, 2, 3},
				"nested": map[string]interface{}{
					"int":           1,
					"ints":          "1,2,3",
					"dontWired":     "123456789",
					"dontWiredTime": "Mon May  6 23:31:43 EDT 2024",
				},
				"Struct": map[string]interface{}{
					"int":  1,
					"ints": "1,2,3",
				},
				"dontWired":     "987654321",
				"dontWiredTime": "Mon May  6 23:31:43 EDT 2024",
				"fight-expansions": []map[string]interface{}{
					{"round": 1, "start-time": "09:00:00", "stop-time": "12:00:00", "min": 10, "max": 100},
					{"round": 2, "start-time": "10:00:00", "stop-time": "13:00:00", "min": 20, "max": 200},
					{"round": 3, "start-time": "11:00:00", "stop-time": "14:00:00", "min": 30, "max": 300},
				},
			},
		}

		expect := NestedStruct{
			CommonStruct: CommonStruct{
				Int:           1,
				Ints:          []int{1, 2, 3},
				Uint:          uint(3),
				Uints:         []uint{1, 2, 3},
				Float:         float64(3),
				Floats:        []float64{1, 2, 3},
				Bool:          true,
				Bools:         []bool{true, false},
				String:        "abc",
				Strings:       []string{"abc", "def", "ghi"},
				Time:          time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
				Duration:      5 * time.Second,
				dontWired:     "",
				dontWiredTime: time.Time{},
			},
			Struct: CommonStruct{
				Int:           1,
				Ints:          []int{1, 2, 3},
				Uint:          uint(3),
				Uints:         []uint{1, 2, 3},
				Float:         float64(3),
				Floats:        []float64{1, 2, 3},
				Bool:          true,
				Bools:         []bool{true, false},
				String:        "abc",
				Strings:       []string{"abc", "def", "ghi"},
				Time:          time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
				Duration:      5 * time.Second,
				dontWired:     "",
				dontWiredTime: time.Time{},
			},
			Nested: CommonStruct{
				Int:           1,
				Ints:          []int{1, 2, 3},
				Uint:          uint(3),
				Uints:         []uint{1, 2, 3},
				Float:         float64(3),
				Floats:        []float64{1, 2, 3},
				Bool:          true,
				Bools:         []bool{true, false},
				String:        "abc",
				Strings:       []string{"abc", "def", "ghi"},
				Time:          time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
				Duration:      5 * time.Second,
				dontWired:     "",
				dontWiredTime: time.Time{},
			},
			Int:      1,
			Ints:     []int{1, 2, 3},
			Uint:     uint(3),
			Uints:    []uint{1, 2, 3},
			Float:    float64(3),
			Floats:   []float64{1, 2, 3},
			Bool:     true,
			Bools:    []bool{true, false},
			String:   "abc",
			Strings:  []string{"abc", "def", "ghi"},
			Time:     time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
			Duration: 5 * time.Second,
			FightExpansions: []FightExpansion{
				{Round: 1, StartTime: "09:00:00", StopTime: "12:00:00", Min: 10, Max: 100},
				{Round: 2, StartTime: "10:00:00", StopTime: "13:00:00", Min: 20, Max: 200},
				{Round: 3, StartTime: "11:00:00", StopTime: "14:00:00", Min: 30, Max: 300},
			},
		}

		tag = Tag("${struct:=}")
		err = assert.Must(Map(m)).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, expect)

		tag = Tag("${struct}")
		err = assert.Must(Map(m)).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, expect)

		err = assert.Must(Map(m["struct"].(map[string]interface{}))).Bind(&s)
		assert.Nil(t, err)
		assert.Equal(t, s, expect)
	})
}

func TestBind_SliceValue(t *testing.T) {

	t.Run("uints", func(t *testing.T) {
		var u []uint

		key := Key("uints")
		err := assert.Must(Map(nil)).Bind(&u, key)
		assert.Error(t, err, "bind \\[\\]uint error: property \"uints\": not exist")

		tag := Tag("${uints:=1,2,3}")
		err = assert.Must(Map(nil)).Bind(&u, tag)
		assert.Nil(t, err)
		assert.Equal(t, u, []uint{1, 2, 3})

		err = assert.Must(Map(map[string]interface{}{
			"uints": "",
		})).Bind(&u, tag)
		assert.Nil(t, err)
		assert.Equal(t, u, []uint{})

		err = assert.Must(Map(map[string]interface{}{
			"uints": 5,
		})).Bind(&u, tag)
		assert.Nil(t, err)
		assert.Equal(t, u, []uint{5})

		err = assert.Must(Map(map[string]interface{}{
			"uints": []uint{5, 6, 7},
		})).Bind(&u, tag)
		assert.Nil(t, err)
		assert.Equal(t, u, []uint{5, 6, 7})

		err = assert.Must(Map(map[string]interface{}{
			"uints": "5, 6, 7",
		})).Bind(&u, tag)
		assert.Nil(t, err)
		assert.Equal(t, u, []uint{5, 6, 7})

		err = assert.Must(Map(map[string]interface{}{
			"uints": "abc",
		})).Bind(&u, tag)
		assert.Error(t, err, "bind \\[\\]uint error: bind \\[\\]uint\\[0\\] error: strconv.ParseUint: parsing \"abc\": invalid syntax")
	})

	t.Run("ints", func(t *testing.T) {
		var i []int

		key := Key("ints")
		err := assert.Must(Map(nil)).Bind(&i, key)
		assert.Error(t, err, "bind \\[\\]int error: property \"ints\": not exist")

		tag := Tag("${ints:=1,2,3}")
		err = assert.Must(Map(nil)).Bind(&i, tag)
		assert.Nil(t, err)
		assert.Equal(t, i, []int{1, 2, 3})

		err = assert.Must(Map(map[string]interface{}{
			"ints": "",
		})).Bind(&i, tag)
		assert.Nil(t, err)
		assert.Equal(t, i, []int{})

		err = assert.Must(Map(map[string]interface{}{
			"ints": 5,
		})).Bind(&i, tag)
		assert.Nil(t, err)
		assert.Equal(t, i, []int{5})

		err = assert.Must(Map(map[string]interface{}{
			"ints": []int{5, 6, 7},
		})).Bind(&i, tag)
		assert.Nil(t, err)
		assert.Equal(t, i, []int{5, 6, 7})

		err = assert.Must(Map(map[string]interface{}{
			"ints": "5, 6, 7",
		})).Bind(&i, tag)
		assert.Nil(t, err)
		assert.Equal(t, i, []int{5, 6, 7})

		err = assert.Must(Map(map[string]interface{}{
			"ints": "abc",
		})).Bind(&i, tag)
		assert.Error(t, err, "bind \\[\\]int error: bind \\[\\]int\\[0\\] error: strconv.ParseInt: parsing \"abc\": invalid syntax")
	})

	t.Run("floats", func(t *testing.T) {
		var f []float32

		key := Key("floats")
		err := assert.Must(Map(nil)).Bind(&f, key)
		assert.Error(t, err, "bind \\[\\]float32 error: property \"floats\": not exist")

		tag := Tag("${floats:=1,2,3}")
		err = assert.Must(Map(nil)).Bind(&f, tag)
		assert.Nil(t, err)
		assert.Equal(t, f, []float32{1, 2, 3})

		err = assert.Must(Map(map[string]interface{}{
			"floats": "",
		})).Bind(&f, tag)
		assert.Nil(t, err)
		assert.Equal(t, f, []float32{})

		err = assert.Must(Map(map[string]interface{}{
			"floats": 5,
		})).Bind(&f, tag)
		assert.Nil(t, err)
		assert.Equal(t, f, []float32{5})

		err = assert.Must(Map(map[string]interface{}{
			"floats": []float32{5, 6, 7},
		})).Bind(&f, tag)
		assert.Nil(t, err)
		assert.Equal(t, f, []float32{5, 6, 7})

		err = assert.Must(Map(map[string]interface{}{
			"floats": "5, 6, 7",
		})).Bind(&f, tag)
		assert.Nil(t, err)
		assert.Equal(t, f, []float32{5, 6, 7})

		err = assert.Must(Map(map[string]interface{}{
			"floats": "abc",
		})).Bind(&f, tag)
		assert.Error(t, err, "bind \\[\\]float32 error: bind \\[\\]float32\\[0\\] error: strconv.ParseFloat: parsing \"abc\": invalid syntax")
	})

	t.Run("bools", func(t *testing.T) {
		var b []bool

		key := Key("bools")
		err := assert.Must(Map(nil)).Bind(&b, key)
		assert.Error(t, err, "bind \\[\\]bool error: property \"bools\": not exist")

		tag := Tag("${bools:=false,true,false}")
		err = assert.Must(Map(nil)).Bind(&b, tag)
		assert.Nil(t, err)
		assert.Equal(t, b, []bool{false, true, false})

		err = assert.Must(Map(map[string]interface{}{
			"bools": "",
		})).Bind(&b, tag)
		assert.Nil(t, err)
		assert.Equal(t, b, []bool{})

		err = assert.Must(Map(map[string]interface{}{
			"bools": true,
		})).Bind(&b, tag)
		assert.Nil(t, err)
		assert.Equal(t, b, []bool{true})

		err = assert.Must(Map(map[string]interface{}{
			"bools": []bool{true, false, true},
		})).Bind(&b, tag)
		assert.Nil(t, err)
		assert.Equal(t, b, []bool{true, false, true})

		err = assert.Must(Map(map[string]interface{}{
			"bools": "true, false, true",
		})).Bind(&b, tag)
		assert.Nil(t, err)
		assert.Equal(t, b, []bool{true, false, true})

		err = assert.Must(Map(map[string]interface{}{
			"bools": "abc",
		})).Bind(&b, tag)
		assert.Error(t, err, "bind \\[\\]bool error: bind \\[\\]bool\\[0\\] error: strconv.ParseBool: parsing \"abc\": invalid syntax")
	})

	t.Run("strings", func(t *testing.T) {
		var s []string

		tag := Tag("${strings:=abc,cde,def}")
		err := assert.Must(Map(nil)).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, []string{"abc", "cde", "def"})

		err = assert.Must(Map(map[string]interface{}{
			"strings": "",
		})).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, []string{})

		err = assert.Must(Map(map[string]interface{}{
			"strings": "def",
		})).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, []string{"def"})

		err = assert.Must(Map(map[string]interface{}{
			"strings": []string{"def", "efg", "ghi"},
		})).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, []string{"def", "efg", "ghi"})
	})

	t.Run("structs", func(t *testing.T) {
		var s []CommonStruct

		tag := Tag("${structs:=abc,cde,def}")
		err := assert.Must(Map(nil)).Bind(&s, tag)
		assert.Error(t, err, "bind \\[\\]conf.CommonStruct error: slice can't have a non empty default value")

		tag = Tag("${structs:=}")
		err = assert.Must(Map(nil)).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, []CommonStruct{})

		tag = Tag("${structs}")
		err = assert.Must(Map(nil)).Bind(&s, tag)
		assert.Error(t, err, "bind \\[\\]conf.CommonStruct error: property \"structs\": not exist")

		err = assert.Must(Map(map[string]interface{}{
			"structs[0]": map[string]interface{}{
				"int":  3,
				"ints": "1,2,3",
			},
			"structs[2]": "",
		})).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, []CommonStruct{
			{
				Int:      3,
				Ints:     []int{1, 2, 3},
				Uint:     uint(3),
				Uints:    []uint{1, 2, 3},
				Float:    float64(3),
				Floats:   []float64{1, 2, 3},
				Bool:     true,
				Bools:    []bool{true, false},
				String:   "abc",
				Strings:  []string{"abc", "def", "ghi"},
				Time:     time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
				Duration: 5 * time.Second,
			},
		})

		err = assert.Must(Map(map[string]interface{}{
			"structs": []interface{}{
				map[string]interface{}{
					"int":  3,
					"ints": "1,2,3",
				},
				map[string]interface{}{
					"int":  3,
					"ints": "1,2,3",
				},
			},
		})).Bind(&s, tag)
		assert.Nil(t, err)
		assert.Equal(t, s, []CommonStruct{
			{
				Int:      3,
				Ints:     []int{1, 2, 3},
				Uint:     uint(3),
				Uints:    []uint{1, 2, 3},
				Float:    float64(3),
				Floats:   []float64{1, 2, 3},
				Bool:     true,
				Bools:    []bool{true, false},
				String:   "abc",
				Strings:  []string{"abc", "def", "ghi"},
				Time:     time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
				Duration: 5 * time.Second,
			},
			{
				Int:      3,
				Ints:     []int{1, 2, 3},
				Uint:     uint(3),
				Uints:    []uint{1, 2, 3},
				Float:    float64(3),
				Floats:   []float64{1, 2, 3},
				Bool:     true,
				Bools:    []bool{true, false},
				String:   "abc",
				Strings:  []string{"abc", "def", "ghi"},
				Time:     time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
				Duration: 5 * time.Second,
			},
		})
	})

	t.Run("structs#2", func(t *testing.T) {
		type AppSecret struct {
			AppId  string `value:"${appid:=}"`
			AppUrl string `value:"${appurl:=}"`
			AppKey string `value:"${appkey:=}"`
		}

		type BaseChannel struct {
			Extends []AppSecret `value:"${extends:=}"`
		}

		type UserPwd struct {
			BaseChannel `value:"${sdk.userpwd}"`
		}

		var channel UserPwd
		err := assert.Must(Map(map[string]interface{}{
			"sdk.userpwd.extends[0].appid":  "app1",
			"sdk.userpwd.extends[0].appurl": "//app1",
			"sdk.userpwd.extends[0].appkey": "app1key",
			"sdk.userpwd.extends[1].appid":  "app2",
			"sdk.userpwd.extends[1].appurl": "//app2",
			"sdk.userpwd.extends[1].appkey": "app2key",
		})).Bind(&channel, Tag("${}"))
		assert.Nil(t, err)
		assert.Equal(t, channel, UserPwd{
			BaseChannel{
				Extends: []AppSecret{
					{AppId: "app1", AppUrl: "//app1", AppKey: "app1key"},
					{AppId: "app2", AppUrl: "//app2", AppKey: "app2key"},
				},
			},
		})
	})
}

func TestBind_MapValue(t *testing.T) {

	t.Run("error#1", func(t *testing.T) {
		var m map[string]uint
		tag := Tag("${map}")
		err := assert.Must(Map(map[string]interface{}{
			"map": "abc",
		})).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]uint error: property 'map' is value")
	})

	t.Run("error#2", func(t *testing.T) {
		var m map[string]uint
		tag := Tag("${map}")
		err := assert.Must(Map(map[string]interface{}{
			"map": map[string]interface{}{
				"a": "1",
				"b": "abc",
			},
		})).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]uint error: bind map\\[string\\]uint error: strconv.ParseUint: parsing \"abc\": invalid syntax")
	})

	t.Run("error#3", func(t *testing.T) {
		var m map[string]uint
		tag := Tag("${map}")
		err := assert.Must(Map(map[string]interface{}{
			"map": []uint{1, 2, 3},
		})).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]uint error: bind map\\[string\\]uint error: property \"map.0\": not exist")
	})

	t.Run("error#4", func(t *testing.T) {
		var v int
		tag := Tag("${none}")
		err := assert.Must(Map(map[string]interface{}{
			"none": []int{1, 2, 3},
		})).Bind(&v, tag)
		assert.Error(t, err, "bind int error: strconv.ParseInt: parsing \"\": invalid syntax")
	})

	t.Run("uint", func(t *testing.T) {
		var m map[string]uint

		tag := Tag("${map:=abc,123}")
		err := assert.Must(Map(nil)).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]uint error: map can't have a non empty default value")

		tag = Tag("${map:=}")
		err = assert.Must(Map(nil)).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]uint{})

		tag = Tag("${map:=}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]uint{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]uint{
			"abc": 1,
			"def": 2,
		})

		tag = Tag("${map}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]uint{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]uint{
			"abc": 1,
			"def": 2,
		})
	})

	t.Run("int", func(t *testing.T) {
		var m map[string]int

		tag := Tag("${map:=abc,123}")
		err := assert.Must(Map(nil)).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]int error: map can't have a non empty default value")

		tag = Tag("${map:=}")
		err = assert.Must(Map(nil)).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]int{})

		tag = Tag("${map:=}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]int{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]int{
			"abc": 1,
			"def": 2,
		})

		tag = Tag("${map}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]int{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]int{
			"abc": 1,
			"def": 2,
		})
	})

	t.Run("float", func(t *testing.T) {
		var m map[string]float32

		tag := Tag("${map:=abc,123}")
		err := assert.Must(Map(nil)).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]float32 error: map can't have a non empty default value")

		tag = Tag("${map:=}")
		err = assert.Must(Map(nil)).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]float32{})

		tag = Tag("${map:=}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]float32{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]float32{
			"abc": 1,
			"def": 2,
		})

		tag = Tag("${map}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]float32{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]float32{
			"abc": 1,
			"def": 2,
		})
	})

	t.Run("string", func(t *testing.T) {
		var m map[string]string

		tag := Tag("${map:=abc,123}")
		err := assert.Must(Map(nil)).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]string error: map can't have a non empty default value")

		tag = Tag("${map:=}")
		err = assert.Must(Map(nil)).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]string{})

		tag = Tag("${map:=}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]float32{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]string{
			"abc": "1",
			"def": "2",
		})

		tag = Tag("${map}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]float32{
				"abc": 1,
				"def": 2,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]string{
			"abc": "1",
			"def": "2",
		})

		tag = Tag("${map.abc}")
		err = assert.Must(Map(map[string]interface{}{
			"map": map[string]float32{
				"abc.a":   1,
				"abc.b":   2,
				"abc.c":   3,
				"abc.d.e": 4,
				"abb.e.f": 5,
				"abc.f.g": 6,
			},
		})).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]string{
			"a":   "1",
			"b":   "2",
			"c":   "3",
			"d.e": "4",
			"f.g": "6",
		})
	})

	t.Run("struct", func(t *testing.T) {
		var m map[string]CommonStruct

		tag := Tag("${map:=abc,123}")
		err := assert.Must(Map(nil)).Bind(&m, tag)
		assert.Error(t, err, "bind map\\[string\\]conf.CommonStruct error: map can't have a non empty default value")

		tag = Tag("${map:=}")
		err = assert.Must(Map(nil)).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, map[string]CommonStruct{})

		input := map[string]interface{}{
			"map": map[string]interface{}{
				"a": map[string]interface{}{
					"int":  3,
					"ints": "1,2,3",
				},
			},
		}

		expect := map[string]CommonStruct{
			"a": {
				Int:      3,
				Ints:     []int{1, 2, 3},
				Uint:     uint(3),
				Uints:    []uint{1, 2, 3},
				Float:    float64(3),
				Floats:   []float64{1, 2, 3},
				Bool:     true,
				Bools:    []bool{true, false},
				String:   "abc",
				Strings:  []string{"abc", "def", "ghi"},
				Time:     time.Date(2017, 6, 17, 13, 20, 15, 0, time.UTC),
				Duration: 5 * time.Second,
			},
		}

		tag = Tag("${map:=}")
		err = assert.Must(Map(input)).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, expect)

		tag = Tag("${map}")
		err = assert.Must(Map(input)).Bind(&m, tag)
		assert.Nil(t, err)
		assert.Equal(t, m, expect)
	})
}

func TestBind_Validate(t *testing.T) {

	t.Run("uint", func(t *testing.T) {
		var v struct {
			Uint uint `value:"${uint:=2}" expr:"$>=3"`
		}

		tag := Key("s")
		err := assert.Must(Map(nil)).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \"\\$\\>\\=3\" for value 2")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"uint": 1,
			},
		})).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \"\\$\\>\\=3\" for value 1")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"uint": 3,
			},
		})).Bind(&v, tag)
		assert.Nil(t, err)
		assert.Equal(t, v.Uint, uint(3))
	})

	t.Run("int", func(t *testing.T) {
		var v struct {
			Int int `value:"${int:=2}" expr:"$>=3"`
		}

		tag := Key("s")
		err := assert.Must(Map(nil)).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \"\\$\\>\\=3\" for value 2")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"int": 1,
			},
		})).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \"\\$\\>\\=3\" for value 1")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"int": 3,
			},
		})).Bind(&v, tag)
		assert.Nil(t, err)
		assert.Equal(t, v.Int, 3)
	})

	t.Run("float", func(t *testing.T) {
		var v struct {
			Float float32 `value:"${float:=2}" expr:"$>=3"`
		}

		tag := Key("s")
		err := assert.Must(Map(nil)).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \"\\$\\>\\=3\" for value 2")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"float": 1,
			},
		})).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \"\\$\\>\\=3\" for value 1")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"float": 3,
			},
		})).Bind(&v, tag)
		assert.Nil(t, err)
		assert.Equal(t, v.Float, float32(3))
	})

	t.Run("string", func(t *testing.T) {
		var v struct {
			String string `value:"${string:=123}" expr:"len($)>=6"`
		}

		tag := Key("s")
		err := assert.Must(Map(nil)).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \\\"len\\(\\$\\)\\>\\=6\\\" for value 123")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"string": "abc",
			},
		})).Bind(&v, tag)
		assert.Error(t, err, "bind .* error: validate failed on \\\"len\\(\\$\\)\\>\\=6\\\" for value abc")

		err = assert.Must(Map(map[string]interface{}{
			"s": map[string]interface{}{
				"string": "123456",
			},
		})).Bind(&v, tag)
		assert.Nil(t, err)
		assert.Equal(t, v.String, "123456")
	})
}

func TestBind_StructValue(t *testing.T) {

	t.Run("unexported", func(t *testing.T) {
		var s struct {
			value int `value:"${a:=3}"`
		}
		err := assert.Must(Map(nil)).Bind(&s)
		assert.Nil(t, err)
		assert.Equal(t, s.value, 3)
	})

	t.Run("error_tag", func(t *testing.T) {
		var s struct {
			value int `value:"a:=3"`
		}
		err := assert.Must(Map(nil)).Bind(&s)
		assert.Error(t, err, "bind .* error: parse tag 'a:=3' error: invalid syntax")
	})
}

func TestBind_StructFilter(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		var s struct {
			Uint uint `value:"${uint:=3}"`
		}
		p := assert.Must(Map(nil))
		v := reflect.ValueOf(&s).Elem()
		param := BindParam{
			Path: v.Type().String(),
		}
		err := param.BindTag("${}", "")
		assert.Nil(t, err)
		filter := func(i interface{}, param BindParam) (bool, error) {
			return false, errors.New("this is an error")
		}
		err = BindValue(p, v, v.Type(), param, filter)
		assert.Error(t, err, "bind .* error: this is an error")
	})

	t.Run("filtered", func(t *testing.T) {
		var s struct {
			Uint uint `value:"${uint:=3}"`
		}
		p := assert.Must(Map(nil))
		v := reflect.ValueOf(&s).Elem()
		param := BindParam{
			Path: v.Type().String(),
		}
		err := param.BindTag("${}", "")
		assert.Nil(t, err)
		filter := func(i interface{}, param BindParam) (bool, error) {
			reflect.ValueOf(i).Elem().SetUint(3)
			return true, nil
		}
		err = BindValue(p, v, v.Type(), param, filter)
		assert.Nil(t, err)
		assert.Equal(t, s.Uint, uint(3))
	})

	t.Run("default", func(t *testing.T) {
		var s struct {
			Uint uint `value:"${uint:=3}"`
		}
		p := assert.Must(Map(nil))
		v := reflect.ValueOf(&s).Elem()
		param := BindParam{
			Path: v.Type().String(),
		}
		err := param.BindTag("${}", "")
		assert.Nil(t, err)
		filter := func(i interface{}, param BindParam) (bool, error) {
			return false, nil
		}
		err = BindValue(p, v, v.Type(), param, filter)
		assert.Nil(t, err)
		assert.Equal(t, s.Uint, uint(3))
	})
}

func TestBind_Splitter(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		name := "splitter"
		RegisterSplitter(name, nil)
		defer RemoveSplitter(name)
		var s []int
		err := assert.Must(Map(map[string]interface{}{
			"s": "1;2;3",
		})).Bind(&s, Tag("${s}||splitter"))
		assert.Error(t, err, "bind \\[\\]int error: error splitter 'splitter'")
	})

	t.Run("error", func(t *testing.T) {
		name := "splitter"
		RegisterSplitter(name, func(s string) ([]string, error) {
			return nil, errors.New("this is an error")
		})
		defer RemoveSplitter(name)
		var s []int
		err := assert.Must(Map(map[string]interface{}{
			"s": "1;2;3",
		})).Bind(&s, Tag("${s}||splitter"))
		assert.Error(t, err, "bind \\[\\]int error: split error: this is an error")
	})

	t.Run("success", func(t *testing.T) {
		name := "splitter"
		RegisterSplitter(name, func(s string) ([]string, error) {
			return strings.Split(s, ";"), nil
		})
		defer RemoveSplitter(name)
		var s []int
		err := assert.Must(Map(map[string]interface{}{
			"s": "1;2;3",
		})).Bind(&s, Tag("${s}||splitter"))
		assert.Nil(t, err)
		assert.Equal(t, s, []int{1, 2, 3})
	})
}

func TestBind_Converter(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		var s struct {
			Point Point `value:"${point}"`
		}
		err := assert.Must(Map(map[string]interface{}{
			"point": "[1,2]",
		})).Bind(&s)
		assert.Error(t, err, "bind .* error: illegal format")
	})

	t.Run("success", func(t *testing.T) {
		var s struct {
			Point Point `value:"${point}"`
		}
		err := assert.Must(Map(map[string]interface{}{
			"point": "(1,2)",
		})).Bind(&s)
		assert.Nil(t, err)
		assert.Equal(t, s.Point, Point{X: 1, Y: 2})
	})
}

func TestBind_ReflectValue(t *testing.T) {

	assert.Panic(t, func() {
		var i int
		v := reflect.ValueOf(i)
		_ = assert.Must(Map(map[string]interface{}{
			"int": 1,
		})).Bind(v, Key("int"))
	}, "reflect: reflect.Value.SetInt using unaddressable value")

	t.Run("error", func(t *testing.T) {
		var i int
		v := reflect.ValueOf(&i)
		err := assert.Must(Map(map[string]interface{}{
			"int": 1,
		})).Bind(v, Key("int"))
		assert.Error(t, err, "bind \\*int error: target should be value type")
	})

	t.Run("success", func(t *testing.T) {
		var i int
		v := reflect.ValueOf(&i).Elem()
		err := assert.Must(Map(map[string]interface{}{
			"int": 1,
		})).Bind(v, Key("int"))
		assert.Nil(t, err)
		assert.Equal(t, i, 1)
	})
}
