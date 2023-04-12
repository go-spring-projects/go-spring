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

package utils

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/limpo1989/go-spring/utils/assert"
)

func TestBeanSelector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	g := NewMockBeanSelector(ctrl)
	g.EXPECT()
}

func TestBeanDefinition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	g := NewMockBeanDefinition(ctrl)
	g.EXPECT().Type().Return(reflect.TypeOf(3))
	assert.Equal(t, g.Type(), reflect.TypeOf(3))
	g.EXPECT().Value().Return(reflect.ValueOf(3))
	assert.Equal(t, g.Value(), reflect.ValueOf(3))
	g.EXPECT().BeanName().Return("")
	assert.Equal(t, g.BeanName(), "")
	g.EXPECT().TypeName().Return("")
	assert.Equal(t, g.TypeName(), "")
	g.EXPECT().ID().Return("")
	assert.Equal(t, g.ID(), "")
	g.EXPECT().Created().Return(false)
	assert.Equal(t, g.Created(), false)
	g.EXPECT().Wired().Return(true)
	assert.Equal(t, g.Wired(), true)
	g.EXPECT().Interface().Return(nil)
	assert.Equal(t, g.Interface(), nil)
}

func TestConverter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	g := NewMockConverter(ctrl)
	g.EXPECT()
}
