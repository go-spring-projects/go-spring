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
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-spring-projects/go-spring/gs/arg"
	"github.com/go-spring-projects/go-spring/gs/cond"
	"github.com/go-spring-projects/go-spring/internal/utils"
)

type beanStatus int8

const (
	Deleted = beanStatus(-1)
	Default = beanStatus(iota)
	Resolving
	Resolved
	Creating
	Created
	Wired
)

func (s beanStatus) String() string {
	switch s {
	case Deleted:
		return "Deleted"
	case Default:
		return "Default"
	case Resolving:
		return "Resolving"
	case Resolved:
		return "Resolved"
	case Creating:
		return "Creating"
	case Created:
		return "Created"
	case Wired:
		return "Wired"
	default:
		return ""
	}
}

func BeanID(typ interface{}, name string) string {
	return utils.TypeName(typ) + ":" + name
}

type BeanInit interface {
	OnInit(ctx context.Context) error
}

type BeanDestroy interface {
	OnDestroy()
}

// BeanDefinition bean.
type BeanDefinition struct {

	// 原始类型的全限定名
	typeName string

	v reflect.Value // 值
	t reflect.Type  // 类型
	f *arg.Callable // 构造函数

	file string // 注册点所在文件
	line int    // 注册点所在行数

	name    string         // 名称
	status  beanStatus     // 状态
	primary bool           // 是否为主版本
	method  bool           // 是否为成员方法
	cond    cond.Condition // 判断条件
	order   float32        // 收集时的顺序
	init    interface{}    // 初始化函数
	destroy interface{}    // 销毁函数
	depends []BeanSelector // 间接依赖项
	exports []reflect.Type // 导出的接口
}

// bdType type of *BeanDefinition
var bdType = reflect.TypeOf((*BeanDefinition)(nil))

// Type Return the type of the bean.
func (d *BeanDefinition) Type() reflect.Type {
	return d.t
}

// Value Return the value of the bean.
func (d *BeanDefinition) Value() reflect.Value {
	return d.v
}

// Interface Return the actual value of the bean.
func (d *BeanDefinition) Interface() interface{} {
	return d.v.Interface()
}

// ID Return the ID of the bean.
func (d *BeanDefinition) ID() string {
	return d.typeName + ":" + d.name
}

// BeanName Return the name of the bean.
func (d *BeanDefinition) BeanName() string {
	return d.name
}

// TypeName Return the fully qualified name of the bean's original type.
func (d *BeanDefinition) TypeName() string {
	return d.typeName
}

// Created Return whether the bean has been created or not.
func (d *BeanDefinition) Created() bool {
	return d.status >= Created
}

// Wired Return whether the bean has been injected or not.
func (d *BeanDefinition) Wired() bool {
	return d.status == Wired
}

// FileLine Return the registration file:line of a bean.
func (d *BeanDefinition) FileLine() string {
	return fmt.Sprintf("%s:%d", d.file, d.line)
}

// getClass Return the type description of a bean.
func (d *BeanDefinition) getClass() string {
	if d.f == nil {
		return "object bean"
	}
	return "constructor bean"
}

func (d *BeanDefinition) String() string {
	return fmt.Sprintf("%s %q %s", d.getClass(), d.ID(), d.FileLine())
}

// Match Test if the fully qualified name of a bean's type and its name both match.
func (d *BeanDefinition) Match(typeName string, beanName string) bool {

	typeIsSame := false
	if typeName == "" || d.typeName == typeName {
		typeIsSame = true
	}

	nameIsSame := false
	if beanName == "" || d.name == beanName {
		nameIsSame = true
	}

	return typeIsSame && nameIsSame
}

// Name Set the bean name.
func (d *BeanDefinition) Name(name string) *BeanDefinition {
	d.name = name
	return d
}

// On Set the condition for a bean.
func (d *BeanDefinition) On(c cond.Condition) *BeanDefinition {
	if nil == d.cond {
		d.cond = c
		return d
	}

	d.cond = cond.Group(cond.And, d.cond, c)
	return d
}

// Order Set the sorting order for a bean, where a smaller value indicates a higher priority or an earlier position in the order.
func (d *BeanDefinition) Order(order float32) *BeanDefinition {
	d.order = order
	return d
}

// DependsOn Set the indirect dependencies for a bean.
func (d *BeanDefinition) DependsOn(selectors ...BeanSelector) *BeanDefinition {
	d.depends = append(d.depends, selectors...)
	return d
}

// Primary mark primary.
func (d *BeanDefinition) Primary() *BeanDefinition {
	d.primary = true
	return d
}

// validLifeCycleFunc 判断是否是合法的用于 bean 生命周期控制的函数，生命周期函数
// 的要求：只能有一个入参并且必须是 bean 的类型，没有返回值或者只返回 error 类型值。
func validLifeCycleFunc(fnType reflect.Type, beanValue reflect.Value) bool {
	if !utils.IsFuncType(fnType) {
		return false
	}

	switch fnType.NumIn() {
	case 1:
		// func(bean)
		// func(bean) error
		if !utils.HasReceiver(fnType, beanValue) {
			return false
		}
	case 2:
		// func(bean, ctx)
		// func(bean, ctx) error
		if !utils.HasReceiver(fnType, beanValue) || !utils.IsContextType(fnType.In(1)) {
			return false
		}
	default:
		return false
	}

	return utils.ReturnNothing(fnType) || utils.ReturnOnlyError(fnType)
}

// Init Set the initialization function for a bean.
func (d *BeanDefinition) Init(fn interface{}) *BeanDefinition {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.init = fn
		return d
	}
	panic(errors.New("init should be func(bean,[ctx]) or func(bean,[ctx])error"))
}

// Destroy Set the destruction function for a bean.
func (d *BeanDefinition) Destroy(fn interface{}) *BeanDefinition {
	if validLifeCycleFunc(reflect.TypeOf(fn), d.Value()) {
		d.destroy = fn
		return d
	}
	panic(errors.New("destroy should be func(bean,[ctx]) or func(bean,[ctx])error"))
}

// Export indicates the types of interface to export.
func (d *BeanDefinition) Export(exports ...interface{}) *BeanDefinition {
	err := d.export(exports...)
	utils.Panic(err).When(err != nil)
	return d
}

// Caller update bean register point on source file:line.
func (d *BeanDefinition) Caller(skip int) *BeanDefinition {
	if _, file, line, ok := runtime.Caller(skip); ok {
		d.file, d.line = file, line
	}
	return d
}

func (d *BeanDefinition) export(exports ...interface{}) error {
	for _, o := range exports {
		var typ reflect.Type
		if t, ok := o.(reflect.Type); ok {
			typ = t
		} else { // 处理 (*error)(nil) 这种导出形式
			typ = utils.Indirect(reflect.TypeOf(o))
		}
		if typ.Kind() != reflect.Interface {
			return errors.New("only interface type can be exported")
		}
		exported := false
		for _, export := range d.exports {
			if typ == export {
				exported = true
				break
			}
		}
		if exported {
			continue
		}
		d.exports = append(d.exports, typ)
	}
	return nil
}

func (d *BeanDefinition) constructor(ctx Context) error {
	if d.init != nil {
		fnValue := reflect.ValueOf(d.init)
		fnValues := []reflect.Value{d.Value()}
		if fnValue.Type().NumIn() > 1 {
			fnValues = append(fnValues, reflect.ValueOf(WithContext(ctx)))
		}

		out := fnValue.Call(fnValues)
		if len(out) > 0 && !out[0].IsNil() {
			return out[0].Interface().(error)
		}
	}

	if f, ok := d.Interface().(BeanInit); ok {
		if err := f.OnInit(WithContext(ctx)); err != nil {
			return err
		}
	}
	return nil
}

func (d *BeanDefinition) destructor() {
	if d.destroy != nil {
		fnValue := reflect.ValueOf(d.destroy)
		fnValues := []reflect.Value{d.Value()}
		if fnValue.Type().NumIn() > 1 {
			fnValues = append(fnValues, reflect.ValueOf(context.Background()))
		}
		fnValue.Call([]reflect.Value{d.Value()})
	}

	if f, ok := d.Interface().(BeanDestroy); ok {
		f.OnDestroy()
	}
}

// NewBean create bean from object or function.
// When registering a regular function, use the form reflect.ValueOf(fn) to avoid conflicts with constructor functions.
func NewBean(objOrCtor interface{}, ctorArgs ...arg.Arg) *BeanDefinition {

	var v reflect.Value
	var fromValue bool
	var method bool
	var name string

	switch i := objOrCtor.(type) {
	case reflect.Value:
		fromValue = true
		v = i
	default:
		v = reflect.ValueOf(i)
	}

	t := v.Type()
	if !utils.IsBeanType(t) {
		panic(errors.New("bean must be ref type"))
	}

	if !v.IsValid() || v.IsNil() {
		panic(errors.New("bean can't be nil"))
	}

	var f *arg.Callable
	_, file, line, _ := runtime.Caller(1)

	// 以 reflect.ValueOf(fn) 形式注册的函数被视为函数对象 bean 。
	if !fromValue && t.Kind() == reflect.Func {

		if !utils.IsConstructor(t) {
			t1 := "func(...)bean"
			t2 := "func(...)(bean, error)"
			panic(fmt.Errorf("constructor should be %s or %s", t1, t2))
		}

		var err error
		f, err = arg.Bind(objOrCtor, ctorArgs, 1)
		utils.Panic(err).When(err != nil)

		out0 := t.Out(0)
		v = reflect.New(out0)
		if utils.IsBeanType(out0) {
			v = v.Elem()
		}

		t = v.Type()
		if !utils.IsBeanType(t) {
			panic(errors.New("bean must be ref type"))
		}

		// 成员方法一般是 xxx/gs.(*Server).Consumer 形式命名
		fnPtr := reflect.ValueOf(objOrCtor).Pointer()
		fnInfo := runtime.FuncForPC(fnPtr)
		funcName := fnInfo.Name()
		name = funcName[strings.LastIndex(funcName, "/")+1:]
		name = name[strings.Index(name, ".")+1:]
		if name[0] == '(' {
			name = name[strings.Index(name, ".")+1:]
		}
		method = strings.LastIndexByte(fnInfo.Name(), ')') > 0
	}

	if t.Kind() == reflect.Ptr && !utils.IsValueType(t.Elem()) {
		panic(errors.New("bean should be *val but not *ref"))
	}

	// Type.String() 一般返回 *pkg.Type 形式的字符串，
	// 我们只取最后的类型名，如有需要请自定义 bean 名称。
	if name == "" {
		s := strings.Split(t.String(), ".")
		name = strings.TrimPrefix(s[len(s)-1], "*")
	}

	return &BeanDefinition{
		t:        t,
		v:        v,
		f:        f,
		name:     name,
		typeName: utils.TypeName(t),
		status:   Default,
		method:   method,
		file:     file,
		line:     line,
	}
}
