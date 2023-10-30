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

package gs

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-spring-projects/go-spring/conf"
	"github.com/go-spring-projects/go-spring/gs/arg"
	"github.com/go-spring-projects/go-spring/internal/utils"
)

func (c *container) Keys() []string {
	return c.p.Keys()
}

func (c *container) Has(key string) bool {
	return c.p.Has(key)
}

func (c *container) Prop(key string, opts ...conf.GetOption) string {
	return c.p.Get(key, opts...)
}

func (c *container) Resolve(s string) (string, error) {
	return c.p.Resolve(s)
}

func (c *container) Bind(i interface{}, args ...conf.BindArg) error {
	return c.p.Bind(i, args...)
}

// Find the bean objects that meet the specified conditions. Note that this function can only guarantee that the returned beans are valid, i.e.,
// not marked for deletion, but it cannot guarantee that property binding and dependency injection have been completed.
func (c *container) Find(selector BeanSelector) ([]utils.BeanDefinition, error) {
	beans, err := c.findBean(selector)
	if err != nil {
		return nil, err
	}
	var ret []utils.BeanDefinition
	for _, b := range beans {
		ret = append(ret, b)
	}
	return ret, nil
}

// Get retrieves the bean objects that meet the specified conditions based on the type and selector.
//
// When i is a receiver of a basic type, it represents that there can only be one bean object that meets the conditions. If none or more than one is found, an error will be returned.
// When i is a receiver of a map type, it represents retrieving any number of bean objects. The keys in the map are the names of the beans, and the values are the addresses of the beans.
// When i is an array or slice, it also represents retrieving any number of bean objects. However, it sorts the retrieved bean objects. If no selector is provided or the selector is "*",
// it sorts the bean objects based on the order value of the beans. This mode is called automatic mode. Otherwise, it sorts the bean objects based on the provided selector list. This mode is called assigned mode.
// The difference between this method and the Find method is that Get guarantees that all returned bean objects have completed property binding and dependency injection, while Find only guarantees that the returned bean objects are valid, i.e., not marked for deletion.
func (c *container) Get(i interface{}, selectors ...BeanSelector) error {

	if i == nil {
		return errors.New("i can't be nil")
	}

	if nil == c.tempContainer {
		return errors.New("Ioc container is auto cleared, if you want use Get/Wire please use autowire tag to inject `gs.Context` ")
	}

	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return errors.New("i must be pointer")
	}

	var tags []wireTag
	for _, s := range selectors {
		tags = append(tags, toWireTag(s))
	}

	stack := newWiringStack(c.logger)
	if err := c.autowire(v.Elem(), tags, false, stack); nil != err || len(stack.beans) > 0 {
		if nil != err {
			err = fmt.Errorf("get bean failed\n%s\n↳%w", stack.path(), err)
		} else {
			err = fmt.Errorf("get bean failed\n%s", stack.path())
		}
		return err
	}
	return nil
}

// Wire If the input to Wire is a bean object, it performs property binding and dependency injection on that bean object.
// If the input is a constructor, it immediately executes that constructor and then performs property binding and dependency injection on the returned result.
// In both cases, the function returns the actual value of the bean object after its execution is complete.
func (c *container) Wire(objOrCtor interface{}, ctorArgs ...arg.Arg) (interface{}, error) {

	if objOrCtor == nil {
		return nil, errors.New("objOrCtor can't be nil")
	}

	if nil == c.tempContainer {
		return nil, errors.New("Ioc container is auto cleared, if you want use Get/Wire please use autowire tag to inject `gs.Context` ")
	}

	b := NewBean(objOrCtor, ctorArgs...)
	stack := newWiringStack(c.logger)
	if err := c.wireBean(b, stack); nil != err || len(stack.beans) > 0 {
		if nil != err {
			err = fmt.Errorf("get bean failed\n%s\n↳%w", stack.path(), err)
		} else {
			err = fmt.Errorf("get bean failed\n%s", stack.path())
		}
		return nil, err
	}
	return b.Interface(), nil
}

func (c *container) Invoke(fn interface{}, args ...arg.Arg) ([]interface{}, error) {

	if !utils.IsFuncType(reflect.TypeOf(fn)) {
		return nil, errors.New("fn should be func type")
	}

	stack := newWiringStack(c.logger)

	defer func() {
		if len(stack.beans) > 0 {
			c.logger.Debug(fmt.Sprintf("wiring path %s", stack.path()))
		}
	}()

	r, err := arg.Bind(fn, args, 1)
	if err != nil {
		return nil, err
	}

	ret, err := r.Call(&argContext{c: c, stack: stack})
	if err != nil {
		return nil, err
	}

	var a []interface{}
	for _, v := range ret {
		a = append(a, v.Interface())
	}
	return a, nil
}
