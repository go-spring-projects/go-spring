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

// Setenv convert property syntax to env.
func Setenv(key string, value string) {
	key = convertToEnv(key)
	err := os.Setenv(key, value)
	utils.Panic(err).When(err != nil)
}

// Run start boot app.
func Run(resourceLocator ...ResourceLocator) error {
	return bootApp.Run(resourceLocator...)
}

// Shutdown close boot app.
func Shutdown(msg ...string) {
	bootApp.Shutdown(msg...)
}

// OnProperty binding a callback when the property key loaded.
func OnProperty(key string, fn interface{}) {
	bootApp.OnProperty(key, fn)
}

// Property set property key/value.
func Property(key string, fn interface{}) {
	bootApp.Property(key, fn)
}

// Accept register bean to Ioc container.
func Accept(b *BeanDefinition) *BeanDefinition {
	return bootApp.container.Accept(b)
}

// Object register bean to Ioc container.
func Object(i interface{}) *BeanDefinition {
	return bootApp.container.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide register bean to Ioc container.
func Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition {
	return bootApp.container.Accept(NewBean(ctor, args...))
}

// Go start a goroutine managed by the IoC container.
func Go(fn func(ctx context.Context)) {
	bootApp.container.Go(fn)
}

// AllowCircularReferences enable circular-references.
func AllowCircularReferences() {
	bootApp.container.AllowCircularReferences()
}
