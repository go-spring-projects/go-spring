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

package cond

import (
	"errors"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils"
	"github.com/go-spring-projects/go-spring/internal/utils/assert"
	"github.com/golang/mock/gomock"
)

func TestOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := NewMockContext(ctrl)
	ok, err := OK().Matches(ctx)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestNot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := NewMockContext(ctrl)
	ok, err := Not(OK()).Matches(ctx)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestOnProperty(t *testing.T) {
	t.Run("no property & no HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := OnProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("has property & no HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ok, err := OnProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("no property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := OnProperty("a", HavingValue("a")).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("diff property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("b")
		ok, err := OnProperty("a", HavingValue("a")).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("same property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("a")
		ok, err := OnProperty("a", HavingValue("a")).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("no property & no HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := OnProperty("a", MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("has property & no HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ok, err := OnProperty("a", MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("no property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := OnProperty("a", HavingValue("a"), MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("diff property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("b")
		ok, err := OnProperty("a", HavingValue("a"), MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("same property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("a")
		ok, err := OnProperty("a", HavingValue("a"), MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("go expression", func(t *testing.T) {
		testcases := []struct {
			propValue    string
			expression   string
			expectResult bool
		}{
			{
				"a",
				"go:$==\"a\"",
				true,
			},
			{
				"a",
				"go:$==\"b\"",
				false,
			},
			{
				"3",
				"go:$==3",
				true,
			},
			{
				"3",
				"go:$==4",
				false,
			},
			{
				"3",
				"go:$>1&&$<5",
				true,
			},
			{
				"false",
				"go:$",
				false,
			},
			{
				"false",
				"go:!$",
				true,
			},
		}
		for _, testcase := range testcases {
			ctrl := gomock.NewController(t)
			ctx := NewMockContext(ctrl)
			ctx.EXPECT().Has("a").Return(true)
			ctx.EXPECT().Prop("a").Return(testcase.propValue)
			ok, err := OnProperty("a", HavingValue(testcase.expression)).Matches(ctx)
			assert.Nil(t, err)
			assert.Equal(t, ok, testcase.expectResult)
			ctrl.Finish()
		}
	})
}

func TestOnMissingProperty(t *testing.T) {
	t.Run("no property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := OnMissingProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("has property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ok, err := OnMissingProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestOnBean(t *testing.T) {
	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, errors.New("error"))
		ok, err := OnBean("a").Matches(ctx)
		assert.Error(t, err, "error")
		assert.False(t, ok)
	})
	t.Run("no bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, nil)
		ok, err := OnBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]BeanDefinition{
			utils.NewMockBeanDefinition(nil),
		}, nil)
		ok, err := OnBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("more than one beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]BeanDefinition{
			utils.NewMockBeanDefinition(nil),
			utils.NewMockBeanDefinition(nil),
		}, nil)
		ok, err := OnBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestOnMissingBean(t *testing.T) {
	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, errors.New("error"))
		ok, err := OnMissingBean("a").Matches(ctx)
		assert.Error(t, err, "error")
		assert.False(t, ok)
	})
	t.Run("no bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, nil)
		ok, err := OnMissingBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]BeanDefinition{
			utils.NewMockBeanDefinition(nil),
		}, nil)
		ok, err := OnMissingBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("more than one beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]BeanDefinition{
			utils.NewMockBeanDefinition(nil),
			utils.NewMockBeanDefinition(nil),
		}, nil)
		ok, err := OnMissingBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestOnSingleBean(t *testing.T) {
	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, errors.New("error"))
		ok, err := OnSingleBean("a").Matches(ctx)
		assert.Error(t, err, "error")
		assert.False(t, ok)
	})
	t.Run("no bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, nil)
		ok, err := OnSingleBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]BeanDefinition{
			utils.NewMockBeanDefinition(nil),
		}, nil)
		ok, err := OnSingleBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("more than one beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]BeanDefinition{
			utils.NewMockBeanDefinition(nil),
			utils.NewMockBeanDefinition(nil),
		}, nil)
		ok, err := OnSingleBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestOnExpression(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := NewMockContext(ctrl)
	ok, err := OnExpression("").Matches(ctx)
	assert.Error(t, err, "unimplemented method")
	assert.False(t, ok)
}

func TestOnMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := NewMockContext(ctrl)
	ok, err := OnMatches(func(ctx Context) (bool, error) {
		return false, nil
	}).Matches(ctx)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestOnProfile(t *testing.T) {
	t.Run("no property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("spring.config.profiles").Return(false)
		ok, err := OnProfile("test").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("diff property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("spring.config.profiles").Return(true)
		ctx.EXPECT().Prop("spring.config.profiles").Return("dev")
		ok, err := OnProfile("test").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("same property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ctx.EXPECT().Has("spring.config.profiles").Return(true)
		ctx.EXPECT().Prop("spring.config.profiles").Return("test")
		ok, err := OnProfile("test").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestConditional(t *testing.T) {
	t.Run("ok && ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := On(OK()).And().Matches(ctx)
		assert.Error(t, err, "no condition in last node")
		assert.False(t, ok)
	})
	t.Run("ok && !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := On(OK()).And().On(Not(OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("ok || ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := On(OK()).Or().Matches(ctx)
		assert.Error(t, err, "no condition in last node")
		assert.False(t, ok)
	})
	t.Run("ok || !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := On(OK()).Or().On(Not(OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestGroup(t *testing.T) {
	t.Run("ok && ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := Group(And, OK()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("ok && !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := Group(And, OK(), Not(OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("ok || ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := Group(Or, OK()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("ok || !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockContext(ctrl)
		ok, err := Group(Or, OK(), Not(OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}
