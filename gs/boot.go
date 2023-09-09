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

package gs

import (
	"context"
	"os"
	"reflect"

	"github.com/limpo1989/go-spring/gs/arg"
	"github.com/limpo1989/go-spring/utils"
)

var bootApp = NewApp()

// Setenv 封装 os.Setenv 函数，如果发生 error 会 panic 。
func Setenv(key string, value string) {
	err := os.Setenv(key, value)
	utils.Panic(err).When(err != nil)
}

// Run 启动程序。
func Run(resourceLocator ...ResourceLocator) error {
	return bootApp.Run(resourceLocator...)
}

// Shutdown 停止程序。
func Shutdown(msg ...string) {
	bootApp.Shutdown(msg...)
}

// OnProperty 注册属性监听
func OnProperty(key string, fn interface{}) {
	bootApp.OnProperty(key, fn)
}

// Property 设置属性键值对
func Property(key string, fn interface{}) {
	bootApp.Property(key, fn)
}

// Accept 注册自定义bean
func Accept(b *BeanDefinition) *BeanDefinition {
	return bootApp.container.Accept(b)
}

// Object 注册一个对象bean
func Object(i interface{}) *BeanDefinition {
	return bootApp.container.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide 注册一个方法bean
func Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition {
	return bootApp.container.Accept(NewBean(ctor, args...))
}

// Go 启动一个受gs管理的协程
func Go(fn func(ctx context.Context)) {
	bootApp.container.Go(fn)
}

// AllowCircularReferences 启用循环依赖（注意构造函数bean循环依赖无解）
func AllowCircularReferences() {
	bootApp.container.AllowCircularReferences()
}
