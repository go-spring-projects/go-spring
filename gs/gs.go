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

// Package gs is a core of Go-Spring, including IOC container, IOC container-based APP.
package gs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-spring-projects/go-spring/conf"
	"github.com/go-spring-projects/go-spring/dync"
	"github.com/go-spring-projects/go-spring/gs/arg"
	"github.com/go-spring-projects/go-spring/gs/cond"
	"github.com/go-spring-projects/go-spring/internal/utils"
)

type refreshState int

const (
	Unrefreshed = refreshState(iota)
	RefreshInit
	Refreshing
	Refreshed
)

var (
	loggerType  = reflect.TypeOf((*Logger)(nil))
	contextType = reflect.TypeOf((*Context)(nil)).Elem()
)

type BeanSelector = utils.BeanSelector

type Container interface {
	Context() context.Context
	Properties() *dync.Properties
	AllowCircularReferences()
	Object(i interface{}) *BeanDefinition
	Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition
	Configuration(i interface{}) *BeanDefinition
	Refresh() error
	Close()
}

// Context provides interfaces for using properties and beans based on reflection after the IoC container is started.
type Context interface {
	Context() context.Context
	Keys() []string
	Has(key string) bool
	Prop(key string, opts ...conf.GetOption) string
	Resolve(s string) (string, error)
	Bind(i interface{}, args ...conf.BindArg) error
	Get(i interface{}, selectors ...BeanSelector) error
	Wire(objOrCtor interface{}, ctorArgs ...arg.Arg) (interface{}, error)
	Invoke(fn interface{}, args ...arg.Arg) ([]interface{}, error)
	Go(fn func(ctx context.Context))
}

// ContextAware injects the Context into a struct as the field GSContext.
type ContextAware struct {
	GSContext Context `autowire:""`
}

type tempContainer struct {
	props           *conf.Properties
	beans           []*BeanDefinition
	beansByName     map[string][]*BeanDefinition
	beansByType     map[reflect.Type][]*BeanDefinition
	mapOfOnProperty map[string]interface{}
}

// The container is the cornerstone of the go-spring framework, implementing the concept of dependency injection mentioned in
// Martin Fowler's article "Inversion of Control Containers and the Dependency Injection pattern."
// However, the original article only refers to dependency injection as the handling of dependencies between objects,
// while some IoC containers, such as Spring, also introduce the handling of properties.
//
// Usually, people use the term "dependency injection" to describe both concepts mentioned above. However,
// it is more appropriate to use the term "property binding" to describe the handling of properties.
// Therefore, go-spring strictly differentiates between these two concepts.
// When describing the handling of beans, either dependency injection or property binding can be used separately,
// or both dependency injection and property binding can be used together.
type container struct {
	*tempContainer
	logger                  *Logger
	ctx                     context.Context
	cancel                  context.CancelFunc
	dependencies            []*BeanDefinition
	state                   refreshState
	wg                      sync.WaitGroup
	p                       *dync.Properties
	contextAware            bool
	allowCircularReferences bool
}

// New make a IoC container.
func New() Container {
	ctx, cancel := context.WithCancel(context.Background())
	return &container{
		ctx:    ctx,
		cancel: cancel,
		p:      dync.New(),
		tempContainer: &tempContainer{
			props:           conf.New(),
			beansByName:     make(map[string][]*BeanDefinition),
			beansByType:     make(map[reflect.Type][]*BeanDefinition),
			mapOfOnProperty: make(map[string]interface{}),
		},
	}
}

// Context return context of IoC container.
func (c *container) Context() context.Context {
	return c.ctx
}

// Properties return properties of IoC container.
func (c *container) Properties() *dync.Properties {
	return c.p
}

func validOnProperty(fn interface{}) error {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return errors.New("fn should be a func(value_type)")
	}
	if t.NumIn() != 1 || !utils.IsValueType(t.In(0)) || t.NumOut() != 0 {
		return errors.New("fn should be a func(value_type)")
	}
	return nil
}

// OnProperty binding a callback when the property key loaded.
func (c *container) OnProperty(key string, fn interface{}) {
	err := validOnProperty(fn)
	utils.Panic(err).When(err != nil)
	c.mapOfOnProperty[key] = fn
}

// Property sets the value of the property corresponding to the key. If the property value for the key already exists, the Set method will overwrite the old value.
// In addition to supporting string type property values, the Set method also supports other primitive data types such as int, uint, bool, and so on for property values.
func (c *container) Property(key string, value interface{}) {
	c.props.Set(key, value)
}

// Accept register bean to Ioc container.
func (c *container) Accept(b *BeanDefinition) *BeanDefinition {
	if c.state >= Refreshing {
		panic(errors.New("should call before Refresh"))
	}
	c.beans = append(c.beans, b)
	return b
}

// Object register object bean to Ioc container.
func (c *container) Object(i interface{}) *BeanDefinition {
	return c.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide register method bean to Ioc container.
func (c *container) Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition {
	return c.Accept(NewBean(ctor, args...))
}

// AllowCircularReferences enable circular-references.
func (c *container) AllowCircularReferences() {
	c.allowCircularReferences = true
}

// Configuration scan that the object `i` has `NewXXX` methods to Ioc container.
//
// example:
//
// func(x *T) NewFoo() *Foo
//
// func(x *T) NewBar(foo *Foo) *Bar
//
// func(x *T) NewSubject(bar *Bar) (*Subject, error)
//
// func(x *T) NewServer() *gs.BeanDefinition
func (c *container) Configuration(i interface{}) *BeanDefinition {

	bValue := reflect.ValueOf(i)
	bType := bValue.Type()

	parentBean := c.Accept(NewBean(bValue))

	for j := 0; j < bType.NumMethod(); j++ {
		method := bType.Method(j)
		if !strings.HasPrefix(method.Name, "New") {
			continue
		}

		var bd *BeanDefinition

		switch method.Type.NumOut() {
		case 1: // func(x *T) NewFoo() *Foo
			outType := method.Type.Out(0)
			if !utils.IsBeanType(outType) {
				continue
			}

			if outType == bdType {
				if method.Type.NumIn() != 1 {
					panic(fmt.Errorf("non-parameter constructor required: %s: %s", method.Name, method.Type.String()))
				}

				bdValues := method.Func.Call([]reflect.Value{bValue})
				bdInst := bdValues[0].Interface()
				bd = c.Accept(bdInst.(*BeanDefinition))
			} else {
				bd = c.Provide(method.Func.Interface())
			}
		case 2:
			out0Type := method.Type.Out(0)
			if !utils.IsBeanType(out0Type) {
				continue
			}

			out1Type := method.Type.Out(1)
			if !utils.IsErrorType(out1Type) {
				panic(fmt.Errorf("%s: %s: second return type must be error", method.Name, method.Type.String()))
			}

			if out0Type == bdType {
				if method.Type.NumIn() != 1 {
					panic(fmt.Errorf("non-parameter constructor required: %s: %s", method.Name, method.Type.String()))
				}

				bdValues := method.Func.Call([]reflect.Value{bValue})
				bdInst := bdValues[0].Interface()
				bdErr := bdValues[1].Interface()
				if err, ok := bdErr.(error); ok && nil != err {
					panic(fmt.Errorf("%s: %w", method.Type.String(), err))
				}
				bd = c.Accept(bdInst.(*BeanDefinition))
			} else {
				bd = c.Provide(method.Func.Interface())
			}
		}

		// 修改注册行号信息为父级bean注册位置
		bd.file, bd.line = parentBean.file, parentBean.line
		// 依赖父级bean
		bd.On(cond.OnBean(parentBean.ID()))
	}

	return parentBean
}

// Dependencies return the dependency order list in either ascending or descending order.
func (c *container) Dependencies(asc bool) (deps []*BeanDefinition) {
	if !asc {
		deps = make([]*BeanDefinition, 0, len(c.dependencies))
		for i := len(c.dependencies) - 1; i >= 0; i-- {
			deps = append(deps, c.dependencies[i])
		}
	} else {
		deps = make([]*BeanDefinition, len(c.dependencies))
		copy(deps[:], c.dependencies[:])
	}
	return
}

type lazyField struct {
	v    reflect.Value
	path string
	tag  string
}

// wiringStack 记录 bean 的注入路径。
type wiringStack struct {
	logger       *Logger
	beans        []*BeanDefinition
	lazyFields   []lazyField
	dependencies []*BeanDefinition
}

func newWiringStack(logger *Logger) *wiringStack {
	return &wiringStack{
		logger: logger,
	}
}

// pushBack 添加一个即将注入的 bean 。
func (s *wiringStack) pushBack(b *BeanDefinition) {
	s.logger.Debug(fmt.Sprintf("push %s %s", b, b.status))
	s.beans = append(s.beans, b)
}

// popBack 删除一个已经注入的 bean 。
func (s *wiringStack) popBack() {
	n := len(s.beans)
	b := s.beans[n-1]
	s.beans = s.beans[:n-1]
	s.logger.Debug(fmt.Sprintf("pop %s %s", b, b.status))
}

// pushDependency 记录依赖顺序
func (s *wiringStack) pushDependency(b *BeanDefinition) {
	for _, bean := range s.dependencies {
		if bean == b {
			return
		}
	}
	s.dependencies = append(s.dependencies, b)
}

// path 返回 bean 的注入路径。
func (s *wiringStack) path() (path string) {
	var sb strings.Builder
	for idx, b := range s.beans {
		if idx > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("↳")
		sb.WriteString(b.String())
	}
	return sb.String()
}

func (c *container) clear() {
	c.tempContainer = nil
}

// Refresh the container's beans, perform validity checks on beans, and complete property binding and dependency injection.
func (c *container) Refresh() error {
	return c.refresh(true)
}

func (c *container) refresh(autoClear bool) (err error) {

	if c.state != Unrefreshed {
		return errors.New("container already refreshed")
	}

	start := time.Now()
	c.state = RefreshInit

	c.Object(c).Export((*Context)(nil))

	for key, f := range c.mapOfOnProperty {
		t := reflect.TypeOf(f)
		in := reflect.New(t.In(0)).Elem()
		if err = c.p.Bind(in, conf.Key(key)); err != nil {
			return err
		}
		reflect.ValueOf(f).Call([]reflect.Value{in})
	}

	c.state = Refreshing
	c.logger = GetLogger("", utils.TypeName(c))

	for _, b := range c.beans {
		c.registerBean(b)
	}

	for _, b := range c.beans {
		if err = c.resolveBean(b); err != nil {
			return err
		}
	}

	beansById := make(map[string]*BeanDefinition)
	{
		for _, b := range c.beans {
			if b.status == Deleted {
				continue
			}
			if b.status != Resolved {
				return fmt.Errorf("unexpected status %d", b.status)
			}
			beanID := b.ID()
			if d, ok := beansById[beanID]; ok {
				return fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
			}
			beansById[beanID] = b
		}
	}

	stack := newWiringStack(c.logger)

	defer func() {
		if err != nil || len(stack.beans) > 0 {
			if nil != err {
				err = fmt.Errorf("container refresh failed\n%s\n↳%w", stack.path(), err)
			} else {
				err = fmt.Errorf("container refresh failed\n%s", stack.path())
			}
		}
	}()

	// 按照 bean id 升序注入，保证注入过程始终一致。
	{
		keys := utils.SortedKeys(beansById)
		for _, s := range keys {
			b := beansById[s]
			if err = c.wireBean(b, stack); err != nil {
				return err
			}
		}
	}

	// 处理被标记为延迟注入的那些 bean 字段
	for _, f := range stack.lazyFields {
		tag := strings.TrimSuffix(f.tag, ",lazy")
		if err := c.wireByTag(f.v, tag, stack); err != nil {
			return err //fmt.Errorf("%q wired error: %s", f.path, err.Error())
		}
	}

	c.dependencies = stack.dependencies
	c.state = Refreshed

	cost := time.Now().Sub(start)
	c.logger.Info(fmt.Sprintf("refresh %d beans cost %v", len(beansById), cost))

	if autoClear && !c.contextAware {
		c.clear()
	}

	c.logger.Info("container refreshed successfully")
	return nil
}

func (c *container) registerBean(b *BeanDefinition) {
	c.logger.Debug(fmt.Sprintf("register %s name:%q type:%q %s", b.getClass(), b.BeanName(), b.Type(), b.FileLine()))
	c.beansByName[b.name] = append(c.beansByName[b.name], b)
	c.beansByType[b.Type()] = append(c.beansByType[b.Type()], b)
	for _, t := range b.exports {
		c.logger.Debug(fmt.Sprintf("register %s name:%q type:%q %s", b.getClass(), b.BeanName(), t, b.FileLine()))
		c.beansByType[t] = append(c.beansByType[t], b)
	}
}

// resolveBean 判断 bean 的有效性，如果 bean 是无效的则被标记为已删除。
func (c *container) resolveBean(b *BeanDefinition) error {

	if b.status >= Resolving {
		return nil
	}

	b.status = Resolving

	// method bean 先确定 parent bean 是否存在
	if b.method {
		selector, ok := b.f.Arg(0)
		if !ok || selector == "" {
			selector, _ = b.f.In(0)
		}
		parents, err := c.findBean(selector)
		if err != nil {
			return err
		}
		n := len(parents)
		if n > 1 {
			msg := fmt.Sprintf("found %d parent beans, bean:%q type:%q [", n, selector, b.t.In(0))
			for _, b := range parents {
				msg += "( " + b.String() + " ), "
			}
			msg = msg[:len(msg)-2] + "]"
			return errors.New(msg)
		} else if n == 0 {
			b.status = Deleted
			return nil
		}
	}

	if b.cond != nil {
		if ok, err := b.cond.Matches(c); err != nil {
			return err
		} else if !ok {
			b.status = Deleted
			return nil
		}
	}

	b.status = Resolved
	return nil
}

// wireTag 注入语法的 tag 分解式，字符串形式的完整格式为 TypeName:BeanName? 。
// 注入语法的字符串表示形式分为三个部分，TypeName 是原始类型的全限定名，BeanName
// 是 bean 注册时设置的名称，? 表示注入结果允许为空。
type wireTag struct {
	typeName string
	beanName string
	nullable bool
}

func parseWireTag(str string) (tag wireTag) {

	if str == "" {
		return
	}

	if n := len(str) - 1; str[n] == '?' {
		tag.nullable = true
		str = str[:n]
	}

	i := strings.Index(str, ":")
	if i < 0 {
		tag.beanName = str
		return
	}

	tag.typeName = str[:i]
	tag.beanName = str[i+1:]
	return
}

func (tag wireTag) String() string {
	b := bytes.NewBuffer(nil)
	if tag.typeName != "" {
		b.WriteString(tag.typeName)
		b.WriteString(":")
	}
	b.WriteString(tag.beanName)
	if tag.nullable {
		b.WriteString("?")
	}
	return b.String()
}

func toWireTag(selector BeanSelector) wireTag {
	switch s := selector.(type) {
	case string:
		return parseWireTag(s)
	case BeanDefinition:
		return parseWireTag(s.ID())
	case *BeanDefinition:
		return parseWireTag(s.ID())
	default:
		return parseWireTag(utils.TypeName(s) + ":")
	}
}

func toWireString(tags []wireTag) string {
	var buf bytes.Buffer
	for i, tag := range tags {
		buf.WriteString(tag.String())
		if i < len(tags)-1 {
			buf.WriteByte(',')
		}
	}
	return buf.String()
}

// findBean 查找符合条件的 bean 对象，注意该函数只能保证返回的 bean 是有效的，
// 即未被标记为删除的，而不能保证已经完成属性绑定和依赖注入。
func (c *container) findBean(selector BeanSelector) ([]*BeanDefinition, error) {

	finder := func(fn func(*BeanDefinition) bool) ([]*BeanDefinition, error) {
		var result []*BeanDefinition
		for _, b := range c.beans {
			if b.status == Resolving || b.status == Deleted || !fn(b) {
				continue
			}
			if err := c.resolveBean(b); err != nil {
				return nil, err
			}
			if b.status == Deleted {
				continue
			}
			result = append(result, b)
		}
		return result, nil
	}

	var t reflect.Type
	switch st := selector.(type) {
	case string, BeanDefinition, *BeanDefinition:
		tag := toWireTag(selector)
		return finder(func(b *BeanDefinition) bool {
			return b.Match(tag.typeName, tag.beanName)
		})
	case reflect.Type:
		t = st
	default:
		t = reflect.TypeOf(st)
	}

	if t.Kind() == reflect.Ptr {
		if e := t.Elem(); e.Kind() == reflect.Interface {
			t = e // 指 (*error)(nil) 形式的 bean 选择器
		}
	}

	return finder(func(b *BeanDefinition) bool {
		if b.Type() == t {
			return true
		}
		for _, typ := range b.exports {
			if typ == t {
				return true
			}
		}
		return false
	})
}

// wireBean 对 bean 进行属性绑定和依赖注入，同时追踪其注入路径。如果 bean 有初始
// 化函数，则在注入完成之后执行其初始化函数。如果 bean 依赖了其他 bean，则首先尝试
// 实例化被依赖的 bean 然后对它们进行注入。
func (c *container) wireBean(b *BeanDefinition, stack *wiringStack) error {

	// bean在决议期间因为条件不满足被删除
	if b.status == Deleted {
		return fmt.Errorf("bean:%q have been deleted", b.ID())
	}

	// 已经注入完成的bean当作成功
	if b.status == Wired {
		stack.pushDependency(b)
		return nil
	}

	// 如果该bean重复被注入说明发生了循环(间接)依赖
	if b.status >= Creating {
		// 对象bean可以部分支持循环引用，前提是开启循环引用支持
		if b.f == nil && c.allowCircularReferences {
			return nil
		}
		// 帮助展示循环依赖栈信息
		stack.pushBack(b)
		return errors.New("found circle autowire")
	}

	b.status = Creating

	stack.pushBack(b)

	// 对当前 bean 的间接依赖项进行注入。
	for _, s := range b.depends {
		beans, err := c.findBean(s)
		if err != nil {
			return err
		}
		for _, d := range beans {
			err = c.wireBean(d, stack)
			if err != nil {
				return err
			}
		}
	}

	v, err := c.getBeanValue(b, stack)
	if err != nil {
		return err
	}

	b.status = Created

	t := v.Type()
	for _, typ := range b.exports {
		if !t.Implements(typ) {
			return fmt.Errorf("%s doesn't implement interface %s", b, typ)
		}
	}

	err = c.wireBeanValue(v, t, stack)
	if err != nil {
		return err
	}

	if err = b.constructor(c); nil != err {
		return err
	}

	b.status = Wired
	stack.popBack()
	stack.pushDependency(b)
	return nil
}

type argContext struct {
	c     *container
	stack *wiringStack
}

func (a *argContext) Matches(c cond.Condition) (bool, error) {
	return c.Matches(a.c)
}

func (a *argContext) Bind(v reflect.Value, tag string) error {
	return a.c.p.Bind(v, conf.Tag(tag))
}

func (a *argContext) Wire(v reflect.Value, tag string) error {
	return a.c.wireByTag(v, tag, a.stack)
}

// getBeanValue 获取 bean 的值，如果是构造函数 bean 则执行其构造函数然后返回执行结果。
func (c *container) getBeanValue(b *BeanDefinition, stack *wiringStack) (reflect.Value, error) {

	if b.f == nil {
		return b.Value(), nil
	}

	out, err := b.f.Call(&argContext{c: c, stack: stack})
	if err != nil {
		return reflect.Value{}, err //fmt.Errorf("%s:%s return error: %w", b.getClass(), b.ID(), err)
	}

	// 构造函数的返回值为值类型时 b.Type() 返回其指针类型。
	if val := out[0]; utils.IsBeanType(val.Type()) {
		// 如果实现接口的是值类型，那么需要转换成指针类型然后再赋值给接口。
		if !val.IsNil() && val.Kind() == reflect.Interface && utils.IsValueType(val.Elem().Type()) {
			v := reflect.New(val.Elem().Type())
			v.Elem().Set(val.Elem())
			b.Value().Set(v)
		} else {
			b.Value().Set(val)
		}
	} else {
		b.Value().Elem().Set(val)
	}

	if b.Value().IsNil() {
		return reflect.Value{}, fmt.Errorf("%s:%q return nil", b.getClass(), b.FileLine())
	}

	v := b.Value()
	// 结果以接口类型返回时需要将原始值取出来才能进行注入。
	if b.Type().Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v, nil
}

// wireBeanValue 对 v 进行属性绑定和依赖注入，v 在传入时应该是一个已经初始化的值。
func (c *container) wireBeanValue(v reflect.Value, t reflect.Type, stack *wiringStack) error {

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// 如整数指针类型的 bean 是无需注入的。
	if v.Kind() != reflect.Struct {
		return nil
	}

	typeName := t.Name()
	if typeName == "" { // 简单类型没有名字
		typeName = t.String()
	}

	param := conf.BindParam{Path: typeName}
	return c.wireStruct(v, t, param, stack)
}

// wireStruct 对结构体进行依赖注入，需要注意的是这里不需要进行属性绑定。
func (c *container) wireStruct(v reflect.Value, t reflect.Type, param conf.BindParam, stack *wiringStack) error {

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		if !fv.CanInterface() {
			fv = utils.PatchValue(fv)
			if !fv.CanInterface() {
				continue
			}
		}

		fieldPath := param.Path + "." + ft.Name

		tag, ok := ft.Tag.Lookup("logger")
		if ok {
			if ft.Type != loggerType {
				return fmt.Errorf("field %s expects type *gs.Logger but got %s", fieldPath, ft.Type.String())
			}

			// tag 预处理，可能通过属性值进行指定。
			if strings.HasPrefix(tag, "${") {
				s, err := c.p.Resolve(tag)
				if nil != err {
					return fmt.Errorf("field %s failed resolve logger name: %w", fieldPath, err)
				}
				tag = s
			} else if tag != "" {
				// use bean name
				tag = parseWireTag(tag).beanName
			}

			l := GetLogger(tag, utils.TypeName(v))
			if nil == l {
				return fmt.Errorf("logger field %s not provide: %s", fieldPath, tag)
			}
			fv.Set(reflect.ValueOf(l))
			continue
		}

		// 支持 autowire 和 inject 两个标签。
		tag, ok = ft.Tag.Lookup("autowire")
		if !ok {
			tag, ok = ft.Tag.Lookup("inject")
		}
		if ok {
			if strings.HasSuffix(tag, ",lazy") {
				if !c.allowCircularReferences {
					return fmt.Errorf("lazy field %s.%s require `AllowCircularReferences`", t.String(), ft.Name)
				}
				f := lazyField{v: fv, path: fieldPath, tag: tag}
				stack.lazyFields = append(stack.lazyFields, f)
			} else {
				if ft.Type == contextType {
					c.contextAware = true
				}
				if err := c.wireByTag(fv, tag, stack); err != nil {
					return err //fmt.Errorf("%q wired error: %w", fieldPath, err)
				}
			}
			continue
		}

		subParam := conf.BindParam{
			Key:  param.Key,
			Path: fieldPath,
		}

		if tag, ok = ft.Tag.Lookup("value"); ok {
			err := subParam.BindTag(tag, ft.Tag)
			if err != nil {
				return err
			}
			if ft.Anonymous {
				err = c.wireStruct(fv, ft.Type, subParam, stack)
				if err != nil {
					return err
				}
			} else {
				err = c.p.BindValue(fv.Addr(), subParam)
				if err != nil {
					return err
				}
			}
			continue
		}

		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			err := c.wireStruct(fv, ft.Type, subParam, stack)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *container) wireByTag(v reflect.Value, tag string, stack *wiringStack) error {

	// tag 预处理，可能通过属性值进行指定。
	if strings.HasPrefix(tag, "${") {
		s, err := c.p.Resolve(tag)
		if err != nil {
			return err
		}
		tag = s
	}

	if tag == "" {
		return c.autowire(v, nil, false, stack)
	}

	var tags []wireTag
	if tag != "?" {
		for _, s := range strings.Split(tag, ",") {
			tags = append(tags, toWireTag(s))
		}
	}
	return c.autowire(v, tags, tag == "?", stack)
}

func (c *container) autowire(v reflect.Value, tags []wireTag, nullable bool, stack *wiringStack) error {
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return c.collectBeans(v, tags, nullable, stack)
	default:
		var tag wireTag
		if len(tags) > 0 {
			tag = tags[0]
		} else if nullable {
			tag.nullable = true
		}
		return c.getBean(v, tag, stack)
	}
}

// getBean 获取 tag 对应的 bean 然后赋值给 v，因此 v 应该是一个未初始化的值。
func (c *container) getBean(v reflect.Value, tag wireTag, stack *wiringStack) error {

	if !v.IsValid() {
		return fmt.Errorf("receiver must be ref type, bean:%q", tag)
	}

	t := v.Type()
	if !utils.IsBeanReceiver(t) {
		return fmt.Errorf("%s is not valid receiver type", t.String())
	}

	var foundBeans []*BeanDefinition
	for _, b := range c.beansByType[t] {
		if b.status == Deleted {
			continue
		}
		if !b.Match(tag.typeName, tag.beanName) {
			continue
		}
		foundBeans = append(foundBeans, b)
	}

	// 指定 bean 名称时通过名称获取，防止未通过 Export 方法导出接口。
	if t.Kind() == reflect.Interface && tag.beanName != "" {
		for _, b := range c.beansByName[tag.beanName] {
			if b.status == Deleted {
				continue
			}
			if !b.Type().AssignableTo(t) {
				continue
			}
			if !b.Match(tag.typeName, tag.beanName) {
				continue
			}

			found := false // 对结果排重
			for _, r := range foundBeans {
				if r == b {
					found = true
					break
				}
			}
			if !found {
				foundBeans = append(foundBeans, b)
				//c.logger.Sugar().Warnf("you should call Export() on %s", b)
			}
		}
	}

	if len(foundBeans) == 0 {
		if tag.nullable {
			return nil
		}
		return fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
	}

	// 优先使用设置成主版本的 bean
	var primaryBeans []*BeanDefinition

	for _, b := range foundBeans {
		if b.primary {
			primaryBeans = append(primaryBeans, b)
		}
	}

	if len(primaryBeans) > 1 {
		msg := fmt.Sprintf("found %d primary beans, bean:%q type:%q [", len(primaryBeans), tag, t)
		for _, b := range primaryBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return errors.New(msg)
	}

	if len(primaryBeans) == 0 && len(foundBeans) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(foundBeans), tag, t)
		for _, b := range foundBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return errors.New(msg)
	}

	var result *BeanDefinition
	if len(primaryBeans) == 1 {
		result = primaryBeans[0]
	} else {
		result = foundBeans[0]
	}

	// 确保找到的 bean 已经完成依赖注入。
	err := c.wireBean(result, stack)
	if err != nil {
		return err
	}

	v.Set(result.Value())
	return nil
}

// filterBean 返回 tag 对应的 bean 在数组中的索引，找不到返回 -1。
func filterBean(beans []*BeanDefinition, tag wireTag, t reflect.Type) (int, error) {

	var found []int
	for i, b := range beans {
		if b.Match(tag.typeName, tag.beanName) {
			found = append(found, i)
		}
	}

	if len(found) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(found), tag, t)
		for _, i := range found {
			msg += "( " + beans[i].String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return -1, errors.New(msg)
	}

	if len(found) > 0 {
		i := found[0]
		return i, nil
	}

	if tag.nullable {
		return -1, nil
	}

	return -1, fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
}

type byOrder []*BeanDefinition

func (b byOrder) Len() int           { return len(b) }
func (b byOrder) Less(i, j int) bool { return b[i].order < b[j].order }
func (b byOrder) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func (c *container) collectBeans(v reflect.Value, tags []wireTag, nullable bool, stack *wiringStack) error {

	t := v.Type()
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Map {
		return fmt.Errorf("should be slice or map in collection mode")
	}

	et := t.Elem()
	if !utils.IsBeanReceiver(et) {
		return fmt.Errorf("%s is not valid receiver type", t.String())
	}

	var beans []*BeanDefinition
	if et.Kind() == reflect.Interface && et.NumMethod() == 0 {
		beans = c.beans
	} else {
		beans = c.beansByType[et]
	}

	{
		var arr []*BeanDefinition
		for _, b := range beans {
			if b.status == Deleted {
				continue
			}
			arr = append(arr, b)
		}
		beans = arr
	}

	if len(tags) > 0 {

		var (
			anyBeans  []*BeanDefinition
			afterAny  []*BeanDefinition
			beforeAny []*BeanDefinition
		)

		foundAny := false
		for _, item := range tags {

			// 是否遇到了"无序"标记
			if item.beanName == "*" {
				if foundAny {
					return fmt.Errorf("more than one * in collection %q", tags)
				}
				foundAny = true
				continue
			}

			index, err := filterBean(beans, item, et)
			if err != nil {
				return err
			}
			if index < 0 {
				continue
			}

			if foundAny {
				afterAny = append(afterAny, beans[index])
			} else {
				beforeAny = append(beforeAny, beans[index])
			}

			tmpBeans := append([]*BeanDefinition{}, beans[:index]...)
			beans = append(tmpBeans, beans[index+1:]...)
		}

		if foundAny {
			anyBeans = append(anyBeans, beans...)
		}

		n := len(beforeAny) + len(anyBeans) + len(afterAny)
		arr := make([]*BeanDefinition, 0, n)
		arr = append(arr, beforeAny...)
		arr = append(arr, anyBeans...)
		arr = append(arr, afterAny...)
		beans = arr
	}

	if len(beans) == 0 && !nullable {
		if len(tags) == 0 {
			return fmt.Errorf("no beans collected for %q", toWireString(tags))
		}
		for _, tag := range tags {
			if !tag.nullable {
				return fmt.Errorf("no beans collected for %q", toWireString(tags))
			}
		}
		return nil
	}

	for _, b := range beans {
		if err := c.wireBean(b, stack); err != nil {
			return err
		}
	}

	var ret reflect.Value
	switch t.Kind() {
	case reflect.Slice:
		sort.Sort(byOrder(beans))
		ret = reflect.MakeSlice(t, 0, 0)
		for _, b := range beans {
			ret = reflect.Append(ret, b.Value())
		}
	case reflect.Map:
		ret = reflect.MakeMap(t)
		for _, b := range beans {
			ret.SetMapIndex(reflect.ValueOf(b.name), b.Value())
		}
	}
	v.Set(ret)
	return nil
}

// Close clos Ioc container.
func (c *container) Close() {
	// send a cancel signal to all coroutines managed by the IoC container and wait for them to complete their exit.
	c.cancel()
	c.wg.Wait()

	c.logger.Info("goroutines exited")

	for _, bean := range c.Dependencies(false) {
		bean.destructor()
	}

	c.logger.Info("container closed")
}

// Go start a goroutine managed by the IoC container.
func (c *container) Go(fn func(ctx context.Context)) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error(fmt.Sprint(r), slog.String("stack", string(debug.Stack())))
			}
		}()
		fn(c.ctx)
	}()
}
