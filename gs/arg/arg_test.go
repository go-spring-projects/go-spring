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

package arg

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"go-spring.dev/spring/gs/cond"
	"go-spring.dev/spring/internal/utils/assert"
)

func TestBind(t *testing.T) {

	t.Run("zero argument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		fn := func() {}
		c, err := Bind(fn, []Arg{}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(values), 0)
	})

	t.Run("one value argument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		expectInt := 0
		fn := func(i int) {
			expectInt = i
		}
		c, err := Bind(fn, []Arg{
			Value(3),
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	t.Run("one ctx value argument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Bind(gomock.Any(), "${a.b.c}").DoAndReturn(func(v, tag interface{}) error {
			v.(reflect.Value).SetInt(3)
			return nil
		})
		expectInt := 0
		fn := func(i int) {
			expectInt = i
		}
		c, err := Bind(fn, []Arg{
			"${a.b.c}",
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	t.Run("one ctx named bean argument", func(t *testing.T) {
		type st struct {
			i int
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Wire(gomock.Any(), "a").DoAndReturn(func(v, tag interface{}) error {
			v.(reflect.Value).Set(reflect.ValueOf(&st{3}))
			return nil
		})
		expectInt := 0
		fn := func(v *st) {
			expectInt = v.i
		}
		c, err := Bind(fn, []Arg{
			"a",
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	t.Run("one ctx unnamed bean argument", func(t *testing.T) {
		type st struct {
			i int
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Wire(gomock.Any(), "").DoAndReturn(func(v, tag interface{}) error {
			v.(reflect.Value).Set(reflect.ValueOf(&st{3}))
			return nil
		})
		expectInt := 0
		fn := func(v *st) {
			expectInt = v.i
		}
		c, err := Bind(fn, []Arg{}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	t.Run("one ctx matches", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Matches(gomock.Any()).Return(true, nil)

		ok, err := ctx.Matches(cond.OK())
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("bind args", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)

		var expectValues = []interface{}{
			0, 1, 2, 3, 4, 5, 6, nil,
		}

		var gotValues []interface{}
		fn := func(r0 int, r1 int, r2 int, r3 int, r4 int, r5 int, r6 int, nilPtr *int) {
			gotValues = append(gotValues, r0)
			gotValues = append(gotValues, r1)
			gotValues = append(gotValues, r2)
			gotValues = append(gotValues, r3)
			gotValues = append(gotValues, r4)
			gotValues = append(gotValues, r5)
			gotValues = append(gotValues, r6)
			gotValues = append(gotValues, nil)
		}

		c, err := Bind(fn, []Arg{
			R0(Value(expectValues[0])),
			R1(Value(expectValues[1])),
			R2(Value(expectValues[2])),
			R3(Value(expectValues[3])),
			R4(Value(expectValues[4])),
			R5(Value(expectValues[5])),
			R6(Value(expectValues[6])),
			Index(7, Nil()),
		}, 1)
		assert.Nil(t, err)

		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(values), 0)
		assert.Equal(t, gotValues, expectValues)

		// mark coverage
		NewMockArg(ctrl).EXPECT()
	})

	t.Run("bind options", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)

		type options struct {
			name string
			age  int
		}

		type fnOption func(opts *options)

		var gotOpts = &options{}

		var optFn = func(option ...fnOption) {
			for _, op := range option {
				op(gotOpts)
			}
		}

		var withName = func(name string) fnOption {
			return func(opts *options) {
				opts.name = name
			}
		}

		var withAge = func(age int) fnOption {
			return func(opts *options) {
				opts.age = age
			}
		}

		c, err := Bind(optFn, []Arg{Option(withName, Value("spring")), Option(withAge, Value(18))}, 1)
		assert.Nil(t, err)

		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(values), 0)
		assert.Equal(t, gotOpts.name, "spring")
		assert.Equal(t, gotOpts.age, 18)

	})
}
