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
	"bytes"
	"errors"
	"fmt"
	"image"
	"log/slog"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-spring-projects/go-spring/conf"
	"github.com/go-spring-projects/go-spring/gs/arg"
	"github.com/go-spring-projects/go-spring/gs/cond"
	pkg1 "github.com/go-spring-projects/go-spring/gs/testdata/pkg/bar"
	pkg2 "github.com/go-spring-projects/go-spring/gs/testdata/pkg/foo"
	"github.com/go-spring-projects/go-spring/internal/utils"
	"github.com/go-spring-projects/go-spring/internal/utils/assert"
	"github.com/spf13/cast"
)

func runTest(c Container, fn func(Context)) error {
	type PandoraAware struct{}
	c.Provide(func(p Context) PandoraAware {
		fn(p)
		return PandoraAware{}
	})
	return c.Refresh()
}

func TestApplicationContext_RegisterBeanFrozen(t *testing.T) {
	assert.Panic(t, func() {
		c := New()
		c.Object(func() {}).Init(func(f func()) {
			c.Object(func() {}) // 不能在这里注册新的 Object
		})
		_ = c.Refresh()
	}, "should call before Refresh")
}

func TestApplicationContext(t *testing.T) {

	/////////////////////////////////////////
	// 自定义数据类型

	t.Run("pkg1.SamePkg", func(t *testing.T) {
		c := New()
		e := pkg1.SamePkg{}

		assert.Panic(t, func() {
			c.Object(e)
		}, "bean must be ref type")

		c.Object(&e)
		c.Object(&e).Name("i3")
		c.Object(&e).Name("i4")

		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("pkg2.SamePkg", func(t *testing.T) {
		c := New()
		e := pkg2.SamePkg{}

		assert.Panic(t, func() {
			c.Object(e)
		}, "bean must be ref type")

		c.Object(&e)
		c.Object(&e).Name("i3")
		c.Object(&e).Name("i4")
		c.Object(&e).Name("i5")

		err := c.Refresh()
		assert.Nil(t, err)
	})
}

type TestBincoreng struct {
	i int
}

func (b *TestBincoreng) String() string {
	if b == nil {
		return ""
	} else {
		return strconv.Itoa(b.i)
	}
}

type TestObject struct {

	// 自定义类型指针
	StructByType *TestBincoreng `inject:"?"`
	StructByName *TestBincoreng `autowire:"struct_ptr?"`

	// 自定义类型指针数组
	StructPtrSliceByType []*TestBincoreng `inject:"?"`
	StructPtrCollection  []*TestBincoreng `autowire:"?"`
	StructPtrSliceByName []*TestBincoreng `autowire:"struct_ptr_slice?"`

	// 接口
	InterfaceByType fmt.Stringer `inject:"?"`
	InterfaceByName fmt.Stringer `autowire:"struct_ptr?"`

	// 接口数组
	InterfaceSliceByType []fmt.Stringer `autowire:"?"`

	InterfaceCollection  []fmt.Stringer `inject:"?"`
	InterfaceCollection2 []fmt.Stringer `autowire:"?"`

	// 指定名称时使用精确匹配模式，不对数组元素进行转换，即便能做到似乎也无意义
	InterfaceSliceByName []fmt.Stringer `autowire:"struct_ptr_slice?"`

	//MapTyType map[string]interface{} `inject:"?"`
	MapByName map[string]interface{} `autowire:"map?"`
	MapByNam2 map[string]interface{} `autowire:"struct_ptr?"`
}

func TestApplicationContext_AutoWireBeans(t *testing.T) {

	c := New()

	obj := &TestObject{}
	c.Object(obj)

	b := TestBincoreng{1}
	c.Object(&b).Name("struct_ptr").Export((*fmt.Stringer)(nil))

	err := runTest(c, func(p Context) {})
	assert.Nil(t, err)

	//assert.Equal(t, len(obj.MapTyType), 4)
	assert.Equal(t, len(obj.MapByName), 0)
	assert.Equal(t, len(obj.MapByNam2), 1)
	fmt.Printf("%+v\n", obj)
}

type SubSubSetting struct {
	Int        int `value:"${int}"`
	DefaultInt int `value:"${default.int:=2}"`
}

type SubSetting struct {
	Int        int `value:"${int}"`
	DefaultInt int `value:"${default.int:=2}"`

	SubSubSetting SubSubSetting `value:"${sub}"`
}

type Setting struct {
	Int        int `value:"${int}"`
	DefaultInt int `value:"${default.int:=2}"`
	// IntPtr     *int `value:"${int}"` // 不支持指针

	Uint        uint `value:"${uint}"`
	DefaultUint uint `value:"${default.uint:=2}"`

	Float        float32 `value:"${float}"`
	DefaultFloat float32 `value:"${default.float:=2}"`

	// Complex complex64 `value:"${complex}"` // 不支持复数

	String        string `value:"${string}"`
	DefaultString string `value:"${default.string:=2}"`

	Bool        bool `value:"${bool}"`
	DefaultBool bool `value:"${default.bool:=false}"`

	SubSetting SubSetting `value:"${sub}"`
	// SubSettingPtr *SubSetting `value:"${sub}"` // 不支持指针

	SubSubSetting SubSubSetting `value:"${sub_sub}"`

	IntSlice    []int    `value:"${int_slice}"`
	StringSlice []string `value:"${string_slice}"`
	// FloatSlice  []float64 `value:"${float_slice}"`
}

func TestApplicationContext_ValueTag(t *testing.T) {
	c := New()
	p := conf.New()

	p.Set("int", int(3))
	p.Set("uint", uint(3))
	p.Set("float", float32(3))
	p.Set("complex", complex(3, 0))
	p.Set("string", "3")
	p.Set("bool", true)

	setting := &Setting{}
	c.Object(setting)

	p.Set("sub.int", int(4))
	p.Set("sub.sub.int", int(5))
	p.Set("sub_sub.int", int(6))

	p.Set("int_slice", []int{1, 2})
	p.Set("string_slice", []string{"1", "2"})
	// p.Set("float_slice", []float64{1, 2})

	err := c.Properties().Refresh(p)
	assert.Nil(t, err)

	err = c.Refresh()
	assert.Nil(t, err)

	fmt.Printf("%+v\n", setting)
}

type GreetingService struct {
}

func (gs *GreetingService) Greeting(name string) string {
	return "hello " + name
}

type PrototypeBean struct {
	Service *GreetingService `autowire:""`
	name    string
	t       time.Time
}

func (p *PrototypeBean) Greeting() string {
	return p.t.Format("15:04:05.000") + " " + p.Service.Greeting(p.name)
}

type PrototypeBeanFactory struct {
	PrototypeBean
}

func (f *PrototypeBeanFactory) New(name string) *PrototypeBean {
	b := f.PrototypeBean
	b.name = name
	b.t = time.Now()
	return &b
}

type PrototypeBeanService struct {
	Provide *PrototypeBeanFactory `autowire:""`
}

func (s *PrototypeBeanService) Service(name string) {
	// 通过 PrototypeBean 的工厂获取新的实例，并且每个实例都有自己的时间戳
	fmt.Println(s.Provide.New(name).Greeting())
}

func TestApplicationContext_PrototypeBean(t *testing.T) {
	c := New()

	greetingService := &GreetingService{}
	c.Object(greetingService)

	s := &PrototypeBeanService{}
	c.Object(s)

	f := &PrototypeBeanFactory{}
	c.Object(f)

	err := c.Refresh()
	assert.Nil(t, err)

	s.Service("Li Lei")
	time.Sleep(50 * time.Millisecond)

	s.Service("Jim Green")
	time.Sleep(50 * time.Millisecond)

	s.Service("Han MeiMei")
}

type EnvEnum string

const ENV EnvEnum = "test"

type EnvEnumBean struct {
	EnvType EnvEnum `value:"${env.type}"`
}

type PointBean struct {
	Point        image.Point   `value:"${point}"`
	DefaultPoint image.Point   `value:"${default_point:=(3,4)}"`
	PointList    []image.Point `value:"${point_list}"`
}

func PointConverter(val string) (image.Point, error) {
	if !(strings.HasPrefix(val, "(") && strings.HasSuffix(val, ")")) {
		return image.Point{}, errors.New("数据格式错误")
	}
	ss := strings.Split(val[1:len(val)-1], ",")
	x := cast.ToInt(ss[0])
	y := cast.ToInt(ss[1])
	return image.Point{X: x, Y: y}, nil
}

type DB struct {
	UserName string `value:"${username}"`
	Password string `value:"${password}"`
	Url      string `value:"${url}"`
	Port     string `value:"${port}"`
	DB       string `value:"${db}"`
}

type DbConfig struct {
	DB []DB `value:"${db}"`
}

func TestApplicationContext_TypeConverter(t *testing.T) {
	c := New()
	p := conf.New()
	{
		m, _ := conf.Load("testdata/config/application.yaml")
		for _, key := range m.Keys() {
			p.Set(key, m.Get(key))
		}
	}

	enumBean := &EnvEnumBean{}
	c.Object(enumBean)

	p.Set("env.type", "test")

	pointBean := &PointBean{}
	c.Object(pointBean)

	conf.RegisterConverter(PointConverter)
	p.Set("point", "(7,5)")

	dbConfig := &DbConfig{}
	c.Object(dbConfig)

	err := c.Properties().Refresh(p)
	assert.Nil(t, err)

	err = c.Refresh()
	assert.Nil(t, err)

	assert.Equal(t, enumBean.EnvType, ENV)

	fmt.Printf("%+v\n", enumBean)
	fmt.Printf("%+v\n", pointBean)

	fmt.Printf("%+v\n", dbConfig)
}

type Grouper interface {
	Group()
}

type MyGrouper struct {
}

func (g *MyGrouper) Group() {

}

type ProxyGrouper struct {
	Grouper `autowire:""`
}

func TestApplicationContext_NestedBean(t *testing.T) {
	c := New()
	c.Object(new(MyGrouper)).Export((*Grouper)(nil))
	c.Object(new(ProxyGrouper))
	err := c.Refresh()
	assert.Nil(t, err)
}

type Pkg interface {
	Package()
}

type SamePkgHolder struct {
	// Pkg `autowire:""` // 这种方式会找到多个符合条件的 Object
	Pkg `autowire:"github.com/go-spring-projects/go-spring/gs/testdata/pkg/bar/pkg.SamePkg:SamePkg"`
}

func TestApplicationContext_SameNameBean(t *testing.T) {
	c := New()
	c.Object(new(SamePkgHolder))
	c.Object(&pkg1.SamePkg{}).Export((*Pkg)(nil))
	c.Object(&pkg2.SamePkg{}).Export((*Pkg)(nil))
	err := c.Refresh()
	assert.Nil(t, err)
}

type DiffPkgOne struct {
}

func (d *DiffPkgOne) Package() {
	fmt.Println("github.com/go-spring-projects/go-spring/gs.DiffPkgOne")
}

type DiffPkgTwo struct {
}

func (d *DiffPkgTwo) Package() {
	fmt.Println("github.com/go-spring-projects/go-spring/gs.DiffPkgTwo")
}

type DiffPkgHolder struct {
	// Pkg `autowire:"same"` // 如果两个 Object 不小心重名了，也会找到多个符合条件的 Object
	Pkg `autowire:"github.com/go-spring-projects/go-spring/gs/gs.DiffPkgTwo:same"`
}

func TestApplicationContext_DiffNameBean(t *testing.T) {
	c := New()
	c.Object(&DiffPkgOne{}).Name("same").Export((*Pkg)(nil))
	c.Object(&DiffPkgTwo{}).Name("same").Export((*Pkg)(nil))
	c.Object(new(DiffPkgHolder))
	err := c.Refresh()
	assert.Nil(t, err)
}

func TestApplicationContext_LoadProperties(t *testing.T) {

	c := New()
	p := conf.New()
	{
		m, _ := conf.Load("testdata/config/application.yaml")
		for _, key := range m.Keys() {
			p.Set(key, m.Get(key))
		}

		m, _ = conf.Load("testdata/config/application.properties")
		for _, key := range m.Keys() {
			p.Set(key, m.Get(key))
		}
	}

	err := c.Properties().Refresh(p)
	assert.Nil(t, err)
	err = runTest(c, func(ctx Context) {
		assert.Equal(t, ctx.Prop("yaml.list[0]"), "1")
		assert.Equal(t, ctx.Prop("yaml.list[1]"), "2")
		assert.Equal(t, ctx.Prop("spring.application.name"), "test")
	})
	assert.Nil(t, err)
}

func TestApplicationContext_Get(t *testing.T) {

	t.Run("panic", func(t *testing.T) {
		c := New()
		err := runTest(c, func(p Context) {
			{
				var s fmt.Stringer
				err := p.Get(s)
				assert.Error(t, err, "i can't be nil")
			}
			{
				var s fmt.Stringer
				err := p.Get(&s)
				assert.Error(t, err, "can't find bean, bean:\"\"")
			}
		})
		assert.Nil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		c := New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo)).Export((*Grouper)(nil))
		err := runTest(c, func(p Context) {

			var two *BeanTwo
			err := p.Get(&two)
			assert.Nil(t, err)

			var grouper Grouper
			err = p.Get(&grouper)
			assert.Nil(t, err)

			err = p.Get(&two, (*BeanTwo)(nil))
			assert.Nil(t, err)

			err = p.Get(&grouper, (*BeanTwo)(nil))
			assert.Nil(t, err)

			err = p.Get(&two)
			assert.Nil(t, err)

			err = p.Get(&grouper)
			assert.Nil(t, err)

			err = p.Get(&two, "BeanTwo")
			assert.Nil(t, err)

			err = p.Get(&grouper, "BeanTwo")
			assert.Nil(t, err)

			err = p.Get(&two, ":BeanTwo")
			assert.Nil(t, err)

			err = p.Get(&grouper, ":BeanTwo")
			assert.Nil(t, err)

			err = p.Get(&two, "github.com/go-spring-projects/go-spring/gs/gs.BeanTwo:BeanTwo")
			assert.Nil(t, err)

			err = p.Get(&grouper, "github.com/go-spring-projects/go-spring/gs/gs.BeanTwo:BeanTwo")
			assert.Nil(t, err)

			err = p.Get(&two, "xxx:BeanTwo")
			assert.Error(t, err, "can't find bean, bean:\"xxx:BeanTwo\"")

			err = p.Get(&grouper, "xxx:BeanTwo")
			assert.Error(t, err, "can't find bean, bean:\"xxx:BeanTwo\"")

			var three *BeanThree
			err = p.Get(&three)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})
}

//func TestApplicationContext_FindByName(t *testing.T) {
//
//	c := runTest(c,func(p gs.Context) {})
//	c.Object(&BeanZero{5})
//	c.Object(new(BeanOne))
//	c.Object(new(BeanTwo))
//	c.Refresh()
//
//	p := <-ch
//
//	b, _ := p.Find("")
//	assert.Equal(t, len(b), 4)
//
//	b, _ = p.Find("BeanTwo")
//	fmt.Println(utils.ToJsonString(b))
//	assert.Equal(t, len(b), 0)
//
//	b, _ = p.Find("BeanTwo")
//	fmt.Println(utils.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find(":BeanTwo")
//	fmt.Println(utils.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find("github.com/go-spring-projects/go-spring/gs/gs.BeanTwo:BeanTwo")
//	fmt.Println(utils.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find("xxx:BeanTwo")
//	fmt.Println(utils.ToJsonString(b))
//	assert.Equal(t, len(b), 0)
//
//	b, _ = p.Find((*BeanTwo)(nil))
//	fmt.Println(utils.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find((*fmt.Stringer)(nil))
//	assert.Equal(t, len(b), 0)
//
//	b, _ = p.Find((*Grouper)(nil))
//	assert.Equal(t, len(b), 0)
//}

type Teacher interface {
	Course() string
}

type historyTeacher struct {
	name string
}

func newHistoryTeacher(name string) *historyTeacher {
	return &historyTeacher{name: name}
}

func newTeacher(course string, name string) Teacher {
	switch course {
	case "history":
		return &historyTeacher{name: name}
	default:
		return nil
	}
}

func (t *historyTeacher) Course() string {
	return "history"
}

type Student struct {
	Teacher Teacher
	Room    string
}

// 入参可以进行注入或者属性绑定，返回值可以是 struct、map、slice、func 等。
func NewStudent(teacher Teacher, room string) Student {
	return Student{
		Teacher: teacher,
		Room:    room,
	}
}

// 入参可以进行注入或者属性绑定，返回值可以是 struct、map、slice、func 等。
func NewPtrStudent(teacher Teacher, room string) *Student {
	return &Student{
		Teacher: teacher,
		Room:    room,
	}
}

func TestApplicationContext_RegisterBeanFn(t *testing.T) {
	c := New()
	p := conf.New()

	p.Set("room", "Class 3 Grade 1")

	// 用接口注册时实际使用的是原始类型
	c.Object(Teacher(newHistoryTeacher(""))).Export((*Teacher)(nil))

	c.Provide(NewStudent, "", "${room}").Name("st1")
	c.Provide(NewPtrStudent, "", "${room}").Name("st2")
	c.Provide(NewStudent, "?", "${room:=https://}").Name("st3")
	c.Provide(NewPtrStudent, "?", "${room:=4567}").Name("st4")

	c.Object(newTeacher("history", "")).Init(func(teacher Teacher) {
		fmt.Println(teacher.Course())
	}).Name("newTeacher")

	err := c.Properties().Refresh(p)
	assert.Nil(t, err)

	err = runTest(c, func(p Context) {

		var st1 *Student
		err = p.Get(&st1, "st1")

		assert.Nil(t, err)
		assert.Equal(t, st1.Room, p.Prop("room"))

		var st2 *Student
		err = p.Get(&st2, "st2")

		assert.Nil(t, err)
		assert.Equal(t, st2.Room, p.Prop("room"))

		fmt.Printf("%x\n", reflect.ValueOf(st1).Pointer())
		fmt.Printf("%x\n", reflect.ValueOf(st2).Pointer())

		var st3 *Student
		err = p.Get(&st3, "st3")

		assert.Nil(t, err)
		assert.Equal(t, st3.Room, p.Prop("room"))

		var st4 *Student
		err = p.Get(&st4, "st4")

		assert.Nil(t, err)
		assert.Equal(t, st4.Room, p.Prop("room"))
	})
	assert.Nil(t, err)
}

func TestApplicationContext_Profile(t *testing.T) {

	t.Run("bean:_c:", func(t *testing.T) {
		c := New()
		c.Object(&BeanZero{5})
		err := runTest(c, func(p Context) {
			var b *BeanZero
			err := p.Get(&b)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	})

	t.Run("bean:_c:test", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("spring.config.profiles", "test")
		c.Object(&BeanZero{5})
		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var b *BeanZero
			err = p.Get(&b)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	})
}

type BeanFour struct{}

func TestApplicationContext_DependsOn(t *testing.T) {

	t.Run("random", func(t *testing.T) {
		c := New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanFour))
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("dependsOn", func(t *testing.T) {

		dependsOn := []BeanSelector{
			(*BeanOne)(nil), // 通过类型定义查找
			"github.com/go-spring-projects/go-spring/gs/gs.BeanZero:BeanZero",
		}

		c := New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanFour)).DependsOn(dependsOn...)
		err := c.Refresh()
		assert.Nil(t, err)
	})
}

func TestApplicationContext_Primary(t *testing.T) {

	t.Run("duplicate", func(t *testing.T) {
		c := New()
		c.Object(&BeanZero{5})
		c.Object(&BeanZero{6})
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo))
		err := c.Refresh()
		assert.Error(t, err, "duplicate beans ")
	})

	t.Run("duplicate", func(t *testing.T) {
		c := New()
		c.Object(&BeanZero{5})
		// primary 是在多个候选 bean 里面选择，而不是允许同名同类型的两个 bean
		c.Object(&BeanZero{6}).Primary()
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo))
		err := c.Refresh()
		assert.Error(t, err, "duplicate beans ")
	})

	t.Run("not primary", func(t *testing.T) {
		c := New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo))
		err := runTest(c, func(p Context) {
			var b *BeanTwo
			err := p.Get(&b)
			assert.Nil(t, err)
			assert.Equal(t, b.One.Zero.Int, 5)
		})
		assert.Nil(t, err)
	})

	t.Run("primary", func(t *testing.T) {
		c := New()
		c.Object(&BeanZero{5})
		c.Object(&BeanZero{6}).Name("zero_6").Primary()
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo))
		err := runTest(c, func(p Context) {
			var b *BeanTwo
			err := p.Get(&b)
			assert.Nil(t, err)
			assert.Equal(t, b.One.Zero.Int, 6)
		})
		assert.Nil(t, err)
	})
}

type FuncObj struct {
	Fn func(int) int `autowire:""`
}

func TestDefaultProperties_WireFunc(t *testing.T) {
	c := New()
	c.Object(func(int) int { return 6 })
	obj := new(FuncObj)
	c.Object(obj)
	err := c.Refresh()
	assert.Nil(t, err)
	i := obj.Fn(3)
	assert.Equal(t, i, 6)
}

type Manager interface {
	Cluster() string
}

func NewManager() Manager {
	return localManager{}
}

func NewManagerRetError() (Manager, error) {
	return localManager{}, fmt.Errorf("error")
}

func NewManagerRetErrorNil() (Manager, error) {
	return localManager{}, nil
}

func NewNullPtrManager() Manager {
	return nil
}

func NewPtrManager() Manager {
	return &localManager{}
}

type localManager struct {
	Version string `value:"${manager.version}"`
}

func (m localManager) Cluster() string {
	return "local"
}

func TestApplicationContext_RegisterBeanFn2(t *testing.T) {

	t.Run("ptr manager", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("manager.version", "1.0.0")
		c.Provide(NewPtrManager)
		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var m Manager
			err := p.Get(&m)
			assert.Nil(t, err)

			// 因为用户是按照接口注册的，所以理论上在依赖
			// 系统中用户并不关心接口对应的真实类型是什么。
			var lm *localManager
			err = p.Get(&lm)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("manager", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("manager.version", "1.0.0")

		bd := c.Provide(NewManager)
		assert.Equal(t, bd.BeanName(), "NewManager")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var m Manager
			err := p.Get(&m)
			assert.Nil(t, err)

			var lm *localManager
			err = p.Get(&lm)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("manager return error", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("manager.version", "1.0.0")
		c.Provide(NewManagerRetError)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Error(t, err, "container refresh failed*")
	})

	t.Run("manager return error nil", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("manager.version", "1.0.0")
		c.Provide(NewManagerRetErrorNil)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("manager return nil", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("manager.version", "1.0.0")
		c.Provide(NewNullPtrManager)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Error(t, err, "return nil")
	})
}

type destroyable interface {
	Init()
	Destroy()
	InitWithError() error
	DestroyWithError() error
}

type callDestroy struct {
	i         int
	inited    bool
	destroyed bool
}

func (d *callDestroy) Init() {
	d.inited = true
}

func (d *callDestroy) Destroy() {
	d.destroyed = true
}

func (d *callDestroy) InitWithError() error {
	if d.i == 0 {
		d.inited = true
		return nil
	}
	return fmt.Errorf("error")
}

func (d *callDestroy) DestroyWithError() error {
	if d.i == 0 {
		d.destroyed = true
		return nil
	}
	return fmt.Errorf("error")
}

type nestedCallDestroy struct {
	callDestroy
}

type nestedDestroyable struct {
	destroyable
}

func TestRegisterBean_InitFunc(t *testing.T) {

	t.Run("call init", func(t *testing.T) {

		c := New()
		c.Object(new(callDestroy)).Init((*callDestroy).Init)
		err := runTest(c, func(p Context) {
			var d *callDestroy
			err := p.Get(&d)
			assert.Nil(t, err)
			assert.True(t, d.inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call init with error", func(t *testing.T) {

		{
			c := New()
			c.Object(&callDestroy{i: 1}).Init((*callDestroy).InitWithError)
			err := c.Refresh()
			assert.Error(t, err, "error")
		}

		c := New()
		p := conf.New()
		p.Set("int", 0)
		c.Object(&callDestroy{}).Init((*callDestroy).InitWithError)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var d *callDestroy
			err = p.Get(&d)
			assert.Nil(t, err)
			assert.True(t, d.inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call interface init", func(t *testing.T) {
		c := New()
		c.Provide(func() destroyable { return new(callDestroy) }).Init(destroyable.Init)
		err := runTest(c, func(p Context) {
			var d destroyable
			err := p.Get(&d)
			assert.Nil(t, err)
			assert.True(t, d.(*callDestroy).inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call interface init with error", func(t *testing.T) {

		{
			c := New()
			c.Provide(func() destroyable { return &callDestroy{i: 1} }).Init(destroyable.InitWithError)
			err := c.Refresh()
			assert.Error(t, err, "error")
		}

		c := New()
		p := conf.New()
		p.Set("int", 0)
		c.Provide(func() destroyable { return &callDestroy{} }).Init(destroyable.InitWithError)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var d destroyable
			err = p.Get(&d)
			assert.Nil(t, err)
			assert.True(t, d.(*callDestroy).inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call nested init", func(t *testing.T) {
		c := New()
		c.Object(new(nestedCallDestroy)).Init((*nestedCallDestroy).Init)
		err := runTest(c, func(p Context) {
			var d *nestedCallDestroy
			err := p.Get(&d)
			assert.Nil(t, err)
			assert.True(t, d.inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call nested interface init", func(t *testing.T) {
		c := New()
		c.Object(&nestedDestroyable{
			destroyable: new(callDestroy),
		}).Init((*nestedDestroyable).Init)
		err := runTest(c, func(p Context) {
			var d *nestedDestroyable
			err := p.Get(&d)
			assert.Nil(t, err)
			assert.True(t, d.destroyable.(*callDestroy).inited)
		})
		assert.Nil(t, err)
	})
}

type RecoresCluster struct {
	Endpoints string `value:"${redis.endpoints}"`
}

func TestApplicationContext_ValueBincoreng(t *testing.T) {
	c := New()
	p := conf.New()
	p.Set("redis.endpoints", "redis://localhost:6379")
	c.Object(new(RecoresCluster))

	err := c.Properties().Refresh(p)
	assert.Nil(t, err)
	err = runTest(c, func(p Context) {
		var cluster *RecoresCluster
		err = p.Get(&cluster)
		fmt.Println(cluster)
		assert.Nil(t, err)
	})
	assert.Nil(t, err)
}

func TestApplicationContext_Collect(t *testing.T) {

	t.Run("", func(t *testing.T) {
		c := New()
		c.Object(&struct {
			Events []ServerInterface `autowire:""`
		}{})
		err := runTest(c, func(ctx Context) {})
		assert.Error(t, err, "no beans collected for \"\"")
	})

	t.Run("", func(t *testing.T) {
		c := New()
		err := runTest(c, func(ctx Context) {
			var Events []ServerInterface
			err := ctx.Get(&Events)
			assert.Error(t, err, "no beans collected for \"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := New()
		c.Object(&struct {
			Events []ServerInterface `autowire:"?"`
		}{})
		err := runTest(c, func(ctx Context) {})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := New()
		err := runTest(c, func(ctx Context) {
			var Events []ServerInterface
			err := ctx.Get(&Events, "?")
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("redis.endpoints", "redis://localhost:6379")
		c.Object(new(RecoresCluster)).Name("one")
		c.Object(new(RecoresCluster))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var rcs []*RecoresCluster
			err = p.Get(&rcs)
			assert.Nil(t, err)
			assert.Equal(t, len(rcs), 2)
		})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {

		c := New()
		p := conf.New()
		p.Set("redis.endpoints", "redis://localhost:6379")
		c.Object(new(RecoresCluster)).Name("a").Order(1)
		c.Object(new(RecoresCluster)).Name("b").Order(2)

		intBean := c.Provide(func(p Context) func() {

			var rcs []*RecoresCluster
			err := p.Get(&rcs)

			assert.Nil(t, err)
			assert.Equal(t, len(rcs), 2)
			assert.Equal(t, rcs[0].Endpoints, "redis://localhost:6379")

			return func() {}
		})
		assert.Equal(t, intBean.BeanName(), "TestApplicationContext_Collect.func6.1")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Nil(t, err)
	})
}

var defaultClassOption = ClassOption{
	className: "default",
}

type ClassBuilder struct {
	param string
}

type ClassOption struct {
	className string
	students  []*Student
	floor     int
	builder   *ClassBuilder
}

type ClassOptionFunc func(opt *ClassOption)

func withClassName(className string, floor int) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.className = className
		opt.floor = floor
	}
}

func withStudents(students []*Student) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.students = students
	}
}

func withBuilder(builder *ClassBuilder) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.builder = builder
	}
}

type ClassRoom struct {
	President string `value:"${president}"`
	className string
	floor     int
	students  []*Student
	desktop   Desktop
	builder   *ClassBuilder
}

type Desktop interface {
}

type MetalDesktop struct {
}

func (cls *ClassRoom) Desktop() Desktop {
	return cls.desktop
}

func NewClassRoom(options ...ClassOptionFunc) ClassRoom {
	opt := defaultClassOption
	for _, fn := range options {
		fn(&opt)
	}
	return ClassRoom{
		className: opt.className,
		students:  opt.students,
		floor:     opt.floor,
		desktop:   &MetalDesktop{},
		builder:   opt.builder,
	}
}

func TestOptionPattern(t *testing.T) {

	students := []*Student{
		new(Student), new(Student),
	}

	cls := NewClassRoom()
	assert.Equal(t, cls.className, "default")

	cls = NewClassRoom(withClassName("二年级03班", 3))
	assert.Equal(t, cls.floor, 3)
	assert.Equal(t, len(cls.students), 0)
	assert.Equal(t, cls.className, "二年级03班")

	cls = NewClassRoom(withStudents(students))
	assert.Equal(t, cls.floor, 0)
	assert.Equal(t, cls.students, students)
	assert.Equal(t, cls.className, "default")

	cls = NewClassRoom(withClassName("二年级03班", 3), withStudents(students))
	assert.Equal(t, cls.className, "二年级03班")
	assert.Equal(t, cls.students, students)
	assert.Equal(t, cls.floor, 3)

	cls = NewClassRoom(withStudents(students), withClassName("二年级03班", 3))
	assert.Equal(t, cls.className, "二年级03班")
	assert.Equal(t, cls.students, students)
	assert.Equal(t, cls.floor, 3)
}

func TestOptionConstructorArg(t *testing.T) {

	t.Run("option default", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("president", "CaiYuanPei")
		c.Provide(NewClassRoom)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var cls *ClassRoom
			err := p.Get(&cls)
			assert.Nil(t, err)
			assert.Equal(t, len(cls.students), 0)
			assert.Equal(t, cls.className, "default")
			assert.Equal(t, cls.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("option withClassName", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("president", "CaiYuanPei")
		c.Provide(NewClassRoom, arg.Option(withClassName, "${class_name:=二年级03班}", "${class_floor:=3}"))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var cls *ClassRoom
			err := p.Get(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.floor, 3)
			assert.Equal(t, len(cls.students), 0)
			assert.Equal(t, cls.className, "二年级03班")
			assert.Equal(t, cls.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("option withStudents", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("class_name", "二年级03班")
		p.Set("president", "CaiYuanPei")
		c.Provide(NewClassRoom, arg.Option(withStudents))
		c.Object(new(Student)).Name("Student1")
		c.Object(new(Student)).Name("Student2")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var cls *ClassRoom
			err = p.Get(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.floor, 0)
			assert.Equal(t, len(cls.students), 2)
			assert.Equal(t, cls.className, "default")
			assert.Equal(t, cls.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("option withStudents withClassName", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("class_name", "二年级06班")
		p.Set("president", "CaiYuanPei")
		c.Provide(NewClassRoom,
			arg.Option(withStudents),
			arg.Option(withClassName, "${class_name:=二年级03班}", "${class_floor:=3}"),
			arg.Option(withBuilder, arg.Provide(func(param string) *ClassBuilder {
				return &ClassBuilder{param: param}
			}, arg.Value("1"))),
		)
		c.Object(&Student{}).Name("Student1")
		c.Object(&Student{}).Name("Student2")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var cls *ClassRoom
			err = p.Get(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.floor, 3)
			assert.Equal(t, len(cls.students), 2)
			assert.Equal(t, cls.className, "二年级06班")
			assert.Equal(t, cls.President, "CaiYuanPei")
			assert.Equal(t, cls.builder.param, "1")
		})
		assert.Nil(t, err)
	})
}

type ServerInterface interface {
	Consumer() *Consumer
	ConsumerT() *Consumer
	ConsumerArg(i int) *Consumer
}

type Server struct {
	Version string `value:"${server.version}"`
}

func NewServerInterface() ServerInterface {
	return new(Server)
}

type Consumer struct {
	s *Server
}

func (s *Server) Consumer() *Consumer {
	if nil == s {
		panic(errors.New("server is nil"))
	}
	return &Consumer{s}
}

func (s *Server) ConsumerT() *Consumer {
	return s.Consumer()
}

func (s *Server) ConsumerArg(_ int) *Consumer {
	if nil == s {
		panic(errors.New("server is nil"))
	}
	return &Consumer{s}
}

type Service struct {
	Consumer *Consumer `autowire:""`
}

func TestApplicationContext_RegisterMethodBean(t *testing.T) {

	t.Run("method bean", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		parent := c.Object(new(Server))
		bd := c.Provide((*Server).Consumer, parent)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var s *Server
			err = p.Get(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Version, "1.0.0")

			s.Version = "2.0.0"

			var consumer *Consumer
			err = p.Get(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
		assert.Equal(t, bd.BeanName(), "Consumer")
	})

	t.Run("method bean condition", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		parent := c.Object(new(Server)).On(cond.Not(cond.OK()))
		bd := c.Provide((*Server).Consumer, parent)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var s *Server
			err = p.Get(&s)
			assert.Error(t, err, "can't find bean, bean:\"\" type:\"\\*gs.Server\"")

			var consumer *Consumer
			err = p.Get(&consumer)
			assert.Error(t, err, "can't find bean, bean:\"\" type:\"\\*gs.Consumer\"")
		})
		assert.Nil(t, err)
		assert.Equal(t, bd.BeanName(), "Consumer")
	})

	t.Run("method bean arg", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		parent := c.Object(new(Server))
		c.Provide((*Server).ConsumerArg, parent, "${i:=9}")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var s *Server
			err = p.Get(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Version, "1.0.0")

			s.Version = "2.0.0"

			var consumer *Consumer
			err = p.Get(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean wire to other bean", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		parent := c.Provide(NewServerInterface)
		c.Provide(ServerInterface.Consumer, parent.ID()).DependsOn("ServerInterface")
		c.Object(new(Service))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var si ServerInterface
			err = p.Get(&si)
			assert.Nil(t, err)

			s := si.(*Server)
			assert.Equal(t, s.Version, "1.0.0")

			s.Version = "2.0.0"

			var consumer *Consumer
			err = p.Get(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("circle autowire", func(t *testing.T) {
		okCount := 0
		errCount := 0
		for i := 0; i < 20; i++ { // 不要排序
			func() {

				defer func() {
					if err := recover(); err != nil {
						errCount++

						var v string
						switch e := err.(type) {
						case error:
							v = e.Error()
						case string:
							v = e
						}

						if !strings.Contains(v, "found circle autowire") {
							panic(errors.New("test error"))
						}
					} else {
						okCount++
					}
				}()

				c := New()
				p := conf.New()
				p.Set("server.version", "1.0.0")
				var server = new(Server)
				var service = new(Service)
				parent := c.Provide(func(svc *Service) *Server {
					if service != svc {
						panic("invalid service")
					}
					return server
				})
				c.Provide((*Server).Consumer, parent.ID())
				c.Object(service)

				err := c.Properties().Refresh(p)
				assert.Nil(t, err)
				err = c.Refresh()
				utils.Panic(err).When(err != nil)
			}()
		}
		fmt.Printf("ok:%d err:%d\n", okCount, errCount)
	})

	t.Run("method bean autowire", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		c.Object(new(Server))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var s *Server
			err = p.Get(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Version, "1.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean selector type", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		c.Object(new(Server))
		c.Provide(func(s *Server) *Consumer { return s.Consumer() }, (*Server)(nil))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var s *Server
			err := p.Get(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Version, "1.0.0")

			s.Version = "2.0.0"

			var consumer *Consumer
			err = p.Get(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean selector type error", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		c.Object(new(Server))
		c.Provide(func(s *Server) *Consumer { return s.Consumer() }, (*int)(nil))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Error(t, err, "can't find bean, bean:\"int:\" type:\"\\*gs.Server\"")
	})

	t.Run("method bean selector beanId", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		c.Object(new(Server))
		c.Provide(func(s *Server) *Consumer { return s.Consumer() }, "Server")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var s *Server
			err = p.Get(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Version, "1.0.0")

			s.Version = "2.0.0"

			var consumer *Consumer
			err = p.Get(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean selector beanId error", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		c.Object(new(Server))
		c.Provide(func(s *Server) *Consumer { return s.Consumer() }, "NULL")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Error(t, err, "can't find bean, bean:\"NULL\" type:\"\\*gs.Server\"")
	})
}

func TestApplicationContext_UserDefinedTypeProperty(t *testing.T) {

	type level int

	var config struct {
		Duration time.Duration `value:"${duration}"`
		Level    level         `value:"${level}"`
		Complex  complex64     // `value:"${complex}"`
	}

	c := New()
	p := conf.New()

	conf.RegisterConverter(func(v string) (level, error) {
		if v == "debug" {
			return 1, nil
		}
		return 0, errors.New("error level")
	})

	p.Set("duration", "1h")
	p.Set("level", "debug")
	p.Set("complex", "1+i")
	c.Object(&config)

	err := c.Properties().Refresh(p)
	assert.Nil(t, err)
	err = c.Refresh()
	assert.Nil(t, err)

	fmt.Printf("%+v\n", config)
}

type CircleA struct {
	B *CircleB `autowire:""`
}

type CircleB struct {
	C *CircleC `autowire:""`
}

type CircleC struct {
	A *CircleA `autowire:""`
}

func TestApplicationContext_CircleAutowire(t *testing.T) {

	// 直接创建的 Object 直接发生循环依赖是没有关系的。
	t.Run("", func(t *testing.T) {
		c := New()
		c.AllowCircularReferences()
		var ca = new(CircleA)
		var cb = new(CircleB)
		var cc = new(CircleC)
		c.Object(ca)
		c.Object(cb)
		c.Object(cc)
		err := c.Refresh()
		assert.Nil(t, err)
		assert.Equal(t, ca.B, cb, "CircleB not equal")
		assert.Equal(t, cb.C, cc, "CircleB not equal")
		assert.Equal(t, cc.A, ca, "CircleB not equal")
	})

	t.Run("", func(t *testing.T) {
		c := New()
		c.AllowCircularReferences()
		c.Object(new(CircleA))
		c.Object(new(CircleB))
		c.Provide(func() *CircleC {
			return new(CircleC)
		})
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := New()
		c.AllowCircularReferences()
		c.Object(new(CircleA))
		c.Provide(func() *CircleB {
			return new(CircleB)
		})
		c.Provide(func() *CircleC {
			return new(CircleC)
		})
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := New()
		c.Provide(func(b *CircleB) *CircleA {
			return new(CircleA)
		})
		c.Provide(func(c *CircleC) *CircleB {
			return new(CircleB)
		})
		c.Provide(func(a *CircleA) *CircleC {
			return new(CircleC)
		})
		err := c.Refresh()
		assert.Error(t, err, "found circle autowire")
	})
}

type VarInterfaceOptionFunc func(opt *VarInterfaceOption)

type VarInterfaceOption struct {
	v []interface{}
}

func withVarInterface(v ...interface{}) VarInterfaceOptionFunc {
	return func(opt *VarInterfaceOption) {
		opt.v = v
	}
}

type VarInterfaceObj struct {
	v []interface{}
}

func NewVarInterfaceObj(options ...VarInterfaceOptionFunc) *VarInterfaceObj {
	opt := new(VarInterfaceOption)
	for _, option := range options {
		option(opt)
	}
	return &VarInterfaceObj{opt.v}
}

type Var struct {
	name string
}

type VarOption struct {
	v []*Var
}

type VarOptionFunc func(opt *VarOption)

func withVar(v ...*Var) VarOptionFunc {
	return func(opt *VarOption) {
		opt.v = v
	}
}

type VarObj struct {
	v []*Var
	s string
}

func NewVarObj(s string, options ...VarOptionFunc) *VarObj {
	opt := new(VarOption)
	for _, option := range options {
		option(opt)
	}
	return &VarObj{opt.v, s}
}

func NewNilVarObj(i interface{}, options ...VarOptionFunc) *VarObj {
	opt := new(VarOption)
	for _, option := range options {
		option(opt)
	}
	return &VarObj{opt.v, fmt.Sprint(i)}
}

func TestApplicationContext_RegisterOptionBean(t *testing.T) {

	t.Run("nil param 0", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("var.obj", "description")
		c.Object(&Var{"v1"}).Name("v1")
		c.Object(&Var{"v2"}).Name("v2")
		c.Provide(NewNilVarObj, arg.Nil())

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var obj *VarObj
			err = p.Get(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.v), 0)
			assert.Equal(t, obj.s, "<nil>")
		})
		assert.Nil(t, err)
	})

	t.Run("variable option param 1", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("var.obj", "description")
		c.Object(&Var{"v1"}).Name("v1")
		c.Object(&Var{"v2"}).Name("v2")
		c.Provide(NewVarObj, "${var.obj}", arg.Option(withVar, "v1"))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var obj *VarObj
			err = p.Get(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.v), 1)
			assert.Equal(t, obj.v[0].name, "v1")
			assert.Equal(t, obj.s, "description")
		})
		assert.Nil(t, err)
	})

	t.Run("variable option param 2", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("var.obj", "description")
		c.Object(&Var{"v1"}).Name("v1")
		c.Object(&Var{"v2"}).Name("v2")
		c.Provide(NewVarObj, arg.Value("description"), arg.Option(withVar, "v1", "v2"))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var obj *VarObj
			err = p.Get(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.v), 2)
			assert.Equal(t, obj.v[0].name, "v1")
			assert.Equal(t, obj.v[1].name, "v2")
			assert.Equal(t, obj.s, "description")
		})
		assert.Nil(t, err)
	})

	t.Run("variable option interface param 1", func(t *testing.T) {
		c := New()
		c.Object(&Var{"v1"}).Name("v1").Export((*interface{})(nil))
		c.Object(&Var{"v2"}).Name("v2").Export((*interface{})(nil))
		c.Provide(NewVarInterfaceObj, arg.Option(withVarInterface, "v1"))
		err := runTest(c, func(p Context) {
			var obj *VarInterfaceObj
			err := p.Get(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.v), 1)
		})
		assert.Nil(t, err)
	})

	t.Run("variable option interface param 1", func(t *testing.T) {
		c := New()
		c.Object(&Var{"v1"}).Name("v1").Export((*interface{})(nil))
		c.Object(&Var{"v2"}).Name("v2").Export((*interface{})(nil))
		c.Provide(NewVarInterfaceObj, arg.Option(withVarInterface, "v1", "v2"))
		err := runTest(c, func(p Context) {
			var obj *VarInterfaceObj
			err := p.Get(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.v), 2)
		})
		assert.Nil(t, err)
	})
}

func TestApplicationContext_Close(t *testing.T) {

	t.Run("destroy type", func(t *testing.T) {

		assert.Panic(t, func() {
			c := New()
			c.Object(func() {}).Destroy(func() {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")

		assert.Panic(t, func() {
			c := New()
			c.Object(func() {}).Destroy(func() int { return 0 })
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")

		assert.Panic(t, func() {
			c := New()
			c.Object(func() {}).Destroy(func(int) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")

		assert.Panic(t, func() {
			c := New()
			c.Object(func() {}).Destroy(func(int, int) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
	})

	t.Run("call destroy fn", func(t *testing.T) {
		called := false

		c := New()
		c.Object(func() {}).Destroy(func(f func()) { called = true })
		err := c.Refresh()
		assert.Nil(t, err)
		c.Close()

		assert.True(t, called)
	})

	t.Run("call destroy", func(t *testing.T) {
		c := New()
		d := new(callDestroy)
		c.Object(d).Destroy((*callDestroy).Destroy)
		err := runTest(c, func(p Context) {
			var d *callDestroy
			err := p.Get(&d)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
		c.Close()
		assert.True(t, d.destroyed)
	})

	t.Run("call destroy with error", func(t *testing.T) {

		// error
		{
			c := New()
			d := &callDestroy{i: 1}
			c.Object(d).Destroy((*callDestroy).DestroyWithError)
			err := runTest(c, func(p Context) {
				var d *callDestroy
				err := p.Get(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			assert.False(t, d.destroyed)
		}

		// nil
		{
			c := New()
			d := &callDestroy{}
			c.Object(d).Destroy((*callDestroy).DestroyWithError)
			err := runTest(c, func(p Context) {
				var d *callDestroy
				err := p.Get(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			assert.True(t, d.destroyed)
		}
	})

	t.Run("call interface destroy", func(t *testing.T) {
		c := New()
		bd := c.Provide(func() destroyable { return new(callDestroy) }).Destroy(destroyable.Destroy)
		err := runTest(c, func(p Context) {
			var d destroyable
			err := p.Get(&d)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
		c.Close()
		d := bd.Interface().(*callDestroy)
		assert.True(t, d.destroyed)
	})

	t.Run("call interface destroy with error", func(t *testing.T) {

		// error
		{
			c := New()
			bd := c.Provide(func() destroyable { return &callDestroy{i: 1} }).Destroy(destroyable.DestroyWithError)
			err := runTest(c, func(p Context) {
				var d destroyable
				err := p.Get(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			d := bd.Interface()
			assert.False(t, d.(*callDestroy).destroyed)
		}

		// nil
		{
			c := New()
			p := conf.New()
			p.Set("int", 0)
			bd := c.Provide(func() destroyable { return &callDestroy{} }).Destroy(destroyable.DestroyWithError)

			err := c.Properties().Refresh(p)
			assert.Nil(t, err)
			err = runTest(c, func(p Context) {
				var d destroyable
				err = p.Get(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			d := bd.Interface()
			assert.True(t, d.(*callDestroy).destroyed)
		}
	})
}

type SubNestedAutowireBean struct {
}

type NestedAutowireBean struct {
	SubNestedAutowireBean
}

type PtrNestedAutowireBean struct {
	*SubNestedAutowireBean
}

func TestApplicationContext_NestedAutowireBean(t *testing.T) {
	c := New()
	c.Object(new(NestedAutowireBean))
	c.Object(&PtrNestedAutowireBean{
		SubNestedAutowireBean: new(SubNestedAutowireBean),
	})
	err := runTest(c, func(p Context) {

		var b *NestedAutowireBean
		err := p.Get(&b)
		assert.Nil(t, err)

		var b0 *PtrNestedAutowireBean
		err = p.Get(&b0)
		assert.Nil(t, err)
	})
	assert.Nil(t, err)
}

type BaseChannel struct {
	AutoCreate bool `value:"${auto-create}"`
	Enable     bool `value:"${enable:=false}"`
}

type WXChannel struct {
	BaseChannel `value:"${sdk.wx}"`
}

type baseChannel struct {
	AutoCreate bool `value:"${auto-create}"`

	// nolint 支持对私有字段注入，但是不推荐！代码扫描请忽略这行。
	enable bool `value:"${enable:=false}"`
}

type wxChannel struct {
	baseChannel `value:"${sdk.wx}"`
}

func TestApplicationContext_NestValueField(t *testing.T) {

	t.Run("private", func(t *testing.T) {

		c := New()
		p := conf.New()

		p.Set("sdk.wx.auto-create", true)
		p.Set("sdk.wx.enable", true)

		c.Object(new(wxChannel))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var channel *wxChannel
			err = p.Get(&channel)
			assert.Nil(t, err)
			assert.Equal(t, channel.enable, true)
			assert.Equal(t, channel.AutoCreate, true)
		})
		assert.Nil(t, err)
	})

	t.Run("public", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("sdk.wx.auto-create", true)
		p.Set("sdk.wx.enable", true)
		c.Object(new(WXChannel))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var channel *WXChannel
			err = p.Get(&channel)
			assert.Nil(t, err)
			assert.True(t, channel.Enable)
			assert.True(t, channel.AutoCreate)
		})
		assert.Nil(t, err)
	})
}

func TestApplicationContext_FnArgCollectBean(t *testing.T) {

	t.Run("interface type", func(t *testing.T) {
		c := New()
		c.Provide(newHistoryTeacher("t1")).Name("t1").Export((*Teacher)(nil))
		c.Provide(newHistoryTeacher("t2")).Name("t2").Export((*Teacher)(nil))
		c.Provide(func(teachers []Teacher) func() {
			names := make([]string, 0)
			for _, teacher := range teachers {
				names = append(names, teacher.(*historyTeacher).name)
			}
			sort.Strings(names)
			assert.Equal(t, names, []string{"t1", "t2"})
			return func() {}
		})
		err := c.Refresh()
		assert.Nil(t, err)
	})
}

type filter interface {
	Filter(input string) string
}

type filterImpl struct {
}

func (_ *filterImpl) Filter(input string) string {
	return input
}

func TestApplicationContext_BeanCache(t *testing.T) {

	t.Run("not implement interface", func(t *testing.T) {
		c := New()
		c.Object(func() {}).Export((*filter)(nil))
		err := c.Refresh()
		assert.Error(t, err, "doesn't implement interface gs.filter")
	})

	t.Run("implement interface", func(t *testing.T) {

		var server struct {
			F1 filter `autowire:"f1"`
			F2 filter `autowire:"f2"`
		}

		c := New()
		c.Provide(func() filter { return new(filterImpl) }).Name("f1")
		c.Object(new(filterImpl)).Export((*filter)(nil)).Name("f2")
		c.Object(&server)

		err := c.Refresh()
		assert.Nil(t, err)
	})
}

type IntInterface interface {
	Value() int
}

type Integer int

func (i Integer) Value() int {
	return int(i)
}

func TestApplicationContext_IntInterface(t *testing.T) {
	c := New()
	c.Provide(func() IntInterface { return Integer(5) })
	err := c.Refresh()
	assert.Nil(t, err)
}

type ArrayProperties struct {
	Int      []int           `value:"${int.array:=}"`
	Int8     []int8          `value:"${int8.array:=}"`
	Int16    []int16         `value:"${int16.array:=}"`
	Int32    []int32         `value:"${int32.array:=}"`
	Int64    []int64         `value:"${int64.array:=}"`
	UInt     []uint          `value:"${uint.array:=}"`
	UInt8    []uint8         `value:"${uint8.array:=}"`
	UInt16   []uint16        `value:"${uint16.array:=}"`
	UInt32   []uint32        `value:"${uint32.array:=}"`
	UInt64   []uint64        `value:"${uint64.array:=}"`
	String   []string        `value:"${string.array:=}"`
	Bool     []bool          `value:"${bool.array:=}"`
	Duration []time.Duration `value:"${duration.array:=}"`
	Time     []time.Time     `value:"${time.array:=}"`
}

func TestApplicationContext_Properties(t *testing.T) {

	t.Run("array properties", func(t *testing.T) {
		b := new(ArrayProperties)
		c := New()
		c.Object(b)
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("map default value ", func(t *testing.T) {

		obj := struct {
			Int  int               `value:"${int:=5}"`
			IntA int               `value:"${int_a:=5}"`
			Map  map[string]string `value:"${map:=}"`
			MapA map[string]string `value:"${map_a:=}"`
		}{}

		c := New()
		p := conf.New()
		p.Set("map_a.nba", "nba")
		p.Set("map_a.cba", "cba")
		p.Set("int_a", "3")
		p.Set("int_b", "4")
		c.Object(&obj)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Nil(t, err)

		assert.Equal(t, obj.Int, 5)
		assert.Equal(t, obj.IntA, 3)
	})
}

type FirstDestroy struct {
	T1 *Second1Destroy `autowire:""`
	T2 *Second2Destroy `autowire:""`
}

type Second1Destroy struct {
	T *ThirdDestroy `autowire:""`
}

type Second2Destroy struct {
	T *ThirdDestroy `autowire:""`
}

type ThirdDestroy struct {
}

func TestApplicationContext_Destroy(t *testing.T) {

	destroyIndex := 0
	destroyArray := []int{0, 0, 0, 0}

	c := New()
	c.Object(new(FirstDestroy)).Destroy(
		func(_ *FirstDestroy) {
			fmt.Println("::FirstDestroy")
			destroyArray[destroyIndex] = 1
			destroyIndex++
		})
	c.Object(new(Second2Destroy)).Destroy(
		func(_ *Second2Destroy) {
			fmt.Println("::Second2Destroy")
			destroyArray[destroyIndex] = 2
			destroyIndex++
		})
	c.Object(new(Second1Destroy)).Destroy(
		func(_ *Second1Destroy) {
			fmt.Println("::Second1Destroy")
			destroyArray[destroyIndex] = 2
			destroyIndex++
		})
	c.Object(new(ThirdDestroy)).Destroy(
		func(_ *ThirdDestroy) {
			fmt.Println("::ThirdDestroy")
			destroyArray[destroyIndex] = 4
			destroyIndex++
		})
	err := c.Refresh()
	assert.Nil(t, err)
	c.Close()

	assert.Equal(t, destroyArray, []int{1, 2, 2, 4})
}

type Obj struct {
	i int
}

type ObjFactory struct{}

func (factory *ObjFactory) NewObj(i int) *Obj { return &Obj{i: i} }

func TestApplicationContext_CreateBean(t *testing.T) {
	c := New()
	c.Object(&ObjFactory{})
	err := runTest(c, func(p Context) {
		b, err := p.Wire((*ObjFactory).NewObj, arg.R1("${i:=5}"))
		fmt.Println(b, err)
	})
	assert.Nil(t, err)
}

func TestDefaultSpringContext(t *testing.T) {

	t.Run("bean:test_ctx:", func(t *testing.T) {

		c := New()

		c.Object(&BeanZero{5}).On(cond.
			OnProfile("test").
			And().
			OnMissingBean("null").
			And().
			On(cond.OK()),
		)

		err := runTest(c, func(p Context) {
			var b *BeanZero
			err := p.Get(&b)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("bean:test_ctx:test", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("spring.config.profiles", "test")
		c.Object(&BeanZero{5}).On(cond.OnProfile("test"))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var b *BeanZero
			err = p.Get(&b)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	})

	t.Run("bean:test_ctx:stable", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("spring.config.profiles", "stable")
		c.Object(&BeanZero{5}).On(cond.OnProfile("test"))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var b *BeanZero
			err = p.Get(&b)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("option withClassName Condition", func(t *testing.T) {

		c := New()
		p := conf.New()
		p.Set("president", "CaiYuanPei")
		p.Set("class_floor", 2)
		c.Provide(NewClassRoom, arg.Option(withClassName,
			"${class_name:=二年级03班}",
			"${class_floor:=3}",
		).On(cond.OnProperty("class_name_enable")))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var cls *ClassRoom
			err = p.Get(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.floor, 0)
			assert.Equal(t, len(cls.students), 0)
			assert.Equal(t, cls.className, "default")
			assert.Equal(t, cls.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("option withClassName Apply", func(t *testing.T) {
		onProperty := cond.OnProperty("class_name_enable")
		c := New()
		p := conf.New()
		p.Set("president", "CaiYuanPei")
		c.Provide(NewClassRoom,
			arg.Option(withClassName,
				"${class_name:=二年级03班}",
				"${class_floor:=3}",
			).On(onProperty),
		)

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			var cls *ClassRoom
			err = p.Get(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.floor, 0)
			assert.Equal(t, len(cls.students), 0)
			assert.Equal(t, cls.className, "default")
			assert.Equal(t, cls.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean cond", func(t *testing.T) {
		c := New()
		p := conf.New()
		p.Set("server.version", "1.0.0")
		parent := c.Object(new(Server))
		c.Provide((*Server).Consumer, parent.ID()).On(cond.OnProperty("consumer.enable"))

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {

			var s *Server
			err = p.Get(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Version, "1.0.0")

			var consumer *Consumer
			err = p.Get(&consumer)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})
}

// TODO 现在的方式父 Bean 不存在子 Bean 创建的时候会报错
//func TestDefaultSpringContext_ParentNotRegister(t *testing.T) {
//
//	c := gs.New()
//	parent := c.Provide(NewServerInterface).On(cond.OnProperty("server.is.nil"))
//	c.Provide(ServerInterface.Consumer, parent.ID())
//
//	c.Refresh()
//
//	var s *Server
//	ok := p.Get(&s)
//	utils.Equal(t, ok, false)
//
//	var c *Consumer
//	ok = p.Get(&c)
//	utils.Equal(t, ok, false)
//}

func TestDefaultSpringContext_ConditionOnBean(t *testing.T) {
	c := New()

	c1 := cond.OnProperty("null", cond.MatchIfMissing()).Or().OnProfile("test")

	c.Object(&BeanZero{5}).On(cond.On(c1).And().OnMissingBean("null"))
	c.Object(new(BeanOne)).On(cond.On(c1).And().OnMissingBean("null"))

	c.Object(new(BeanTwo)).On(cond.OnBean("BeanOne"))
	c.Object(new(BeanTwo)).Name("another_two").On(cond.OnBean("Null"))

	err := runTest(c, func(p Context) {

		var two *BeanTwo
		err := p.Get(&two)
		assert.Nil(t, err)

		err = p.Get(&two, "another_two")
		assert.Error(t, err, "can't find bean, bean:\"another_two\"")
	})
	assert.Nil(t, err)
}

func TestDefaultSpringContext_ConditionOnMissingBean(t *testing.T) {
	for i := 0; i < 20; i++ { // 测试 Find 无需绑定，不要排序
		c := New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo)).On(cond.OnMissingBean("BeanOne"))
		c.Object(new(BeanTwo)).Name("another_two").On(cond.OnMissingBean("Null"))
		err := runTest(c, func(p Context) {

			var two *BeanTwo
			err := p.Get(&two)
			assert.Nil(t, err)

			err = p.Get(&two, "another_two")
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	}
}

//func TestFunctionCondition(t *testing.T) {
//	c := gs.New()
//
//	fn := func(c cond.Context) bool { return true }
//	c1 := cond.OnMatches(fn)
//	assert.True(t, c1.Matches(c))
//
//	fn = func(c cond.Context) bool { return false }
//	c2 := cond.OnMatches(fn)
//	assert.False(t, c2.Matches(c))
//}
//
//func TestPropertyCondition(t *testing.T) {
//
//	c := gs.New()
//	p.Set("int", 3)
//	p.Set("parent.child", 0)
//
//	c1 := cond.OnProperty("int")
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnProperty("bool")
//	assert.False(t, c2.Matches(c))
//
//	c3 := cond.OnProperty("parent")
//	assert.True(t, c3.Matches(c))
//
//	c4 := cond.OnProperty("parent123")
//	assert.False(t, c4.Matches(c))
//}
//
//func TestMissingPropertyCondition(t *testing.T) {
//
//	c := gs.New()
//	p.Set("int", 3)
//	p.Set("parent.child", 0)
//
//	c1 := cond.OnMissingProperty("int")
//	assert.False(t, c1.Matches(c))
//
//	c2 := cond.OnMissingProperty("bool")
//	assert.True(t, c2.Matches(c))
//
//	c3 := cond.OnMissingProperty("parent")
//	assert.False(t, c3.Matches(c))
//
//	c4 := cond.OnMissingProperty("parent123")
//	assert.True(t, c4.Matches(c))
//}
//
//func TestPropertyValueCondition(t *testing.T) {
//
//	c := gs.New()
//	p.Set("str", "this is a str")
//	p.Set("int", 3)
//
//	c1 := cond.OnPropertyValue("int", 3)
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnPropertyValue("int", "3")
//	assert.False(t, c2.Matches(c))
//
//	c3 := cond.OnPropertyValue("int", "go:$>2&&$<4")
//	assert.True(t, c3.Matches(c))
//
//	c4 := cond.OnPropertyValue("bool", true)
//	assert.False(t, c4.Matches(c))
//
//	c5 := cond.OnPropertyValue("str", "\"$\"==\"this is a str\"")
//	assert.True(t, c5.Matches(c))
//}
//
//func TestBeanCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Object(&BeanZero{5})
//	c.Object(new(BeanOne))
//	c.Refresh()
//
//	c1 := cond.OnBean("BeanOne")
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnBean("Null")
//	assert.False(t, c2.Matches(c))
//}
//
//func TestMissingBeanCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Object(&BeanZero{5})
//	c.Object(new(BeanOne))
//	c.Refresh()
//
//	c1 := cond.OnMissingBean("BeanOne")
//	assert.False(t, c1.Matches(c))
//
//	c2 := cond.OnMissingBean("Null")
//	assert.True(t, c2.Matches(c))
//}
//
//func TestExpressionCondition(t *testing.T) {
//
//}
//
//func TestConditional(t *testing.T) {
//
//	c := gs.New()
//	p.Set("bool", false)
//	p.Set("int", 3)
//	c.Refresh()
//
//	c1 := cond.OnProperty("int")
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnProperty("int").OnBean("null")
//	assert.False(t, c2.Matches(c))
//
//	assert.Panic(t, func() {
//		c3 := cond.OnProperty("int").And()
//		assert.Equal(t, c3.Matches(c), true)
//	}, "no condition in last node")
//
//	c4 := cond.OnPropertyValue("int", 3).
//		And().
//		OnPropertyValue("bool", false)
//	assert.True(t, c4.Matches(c))
//
//	c5 := cond.OnPropertyValue("int", 3).
//		And().
//		OnPropertyValue("bool", true)
//	assert.False(t, c5.Matches(c))
//
//	c6 := cond.OnPropertyValue("int", 2).
//		Or().
//		OnPropertyValue("bool", true)
//	assert.False(t, c6.Matches(c))
//
//	c7 := cond.OnPropertyValue("int", 2).
//		Or().
//		OnPropertyValue("bool", false)
//	assert.True(t, c7.Matches(c))
//
//	assert.Panic(t, func() {
//		c8 := cond.OnPropertyValue("int", 2).
//			Or().
//			OnPropertyValue("bool", false).
//			Or()
//		assert.Equal(t, c8.Matches(c), true)
//	}, "no condition in last node")
//
//	c9 := cond.OnPropertyValue("int", 2).
//		Or().
//		OnPropertyValue("bool", false).
//		OnPropertyValue("bool", false)
//	assert.True(t, c9.Matches(c))
//}
//
//func TestNotCondition(t *testing.T) {
//
//	c := gs.New()
//	p.Set(environ.SpringProfilesActive, "test")
//	c.Refresh()
//
//	profileCond := cond.OnProfile("test")
//	assert.True(t, profileCond.Matches(c))
//
//	notCond := cond.Not(profileCond)
//	assert.False(t, notCond.Matches(c))
//
//	c1 := cond.OnPropertyValue("int", 2).
//		And().
//		On(cond.Not(profileCond))
//	assert.False(t, c1.Matches(c))
//
//	c2 := cond.OnProfile("test").
//		And().
//		On(cond.Not(profileCond))
//	assert.False(t, c2.Matches(c))
//}

func TestApplicationContext_Invoke(t *testing.T) {

	t.Run("not run", func(t *testing.T) {
		c := New()
		p := conf.New()
		c.Object(func() {})
		p.Set("version", "v0.0.1")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			_, _ = p.Invoke(func(f func(), version string) {
				fmt.Println("version:", version)
			}, "", "${version}")
		})
		assert.Nil(t, err)
	})

	t.Run("run", func(t *testing.T) {
		c := New()
		p := conf.New()
		c.Object(func() {})
		p.Set("version", "v0.0.1")
		p.Set("spring.config.profiles", "dev")

		err := c.Properties().Refresh(p)
		assert.Nil(t, err)
		err = runTest(c, func(p Context) {
			fn := func(f func(), version string) {
				fmt.Println("version:", version)
			}
			_, _ = p.Invoke(fn, "", "${version}")
		})
		assert.Nil(t, err)
	})
}

type emptyStructA struct{}

type emptyStructB struct{}

func TestEmptyStruct(t *testing.T) {

	c := New()
	objA := &emptyStructA{}
	c.Object(objA)
	objB := &emptyStructB{}
	c.Object(objB)
	err := c.Refresh()
	assert.Nil(t, err)

	// objA 和 objB 的地址相同但是类型确实不一样。
	fmt.Printf("objA:%p objB:%p\n", objA, objB)
	fmt.Printf("objA:%#v objB:%#v\n", objA, objB)
}

func TestMapCollection(t *testing.T) {

	type mapValue struct {
		v string
	}

	t.Run("", func(t *testing.T) {
		c := New()
		c.Object(&mapValue{"a"}).Name("a").Order(1)
		c.Object(&mapValue{"b"}).Name("b").Order(2)
		c.Object(&mapValue{"c"}).Name("c").On(cond.Not(cond.OK()))
		err := runTest(c, func(p Context) {

			var vSlice []*mapValue
			err := p.Get(&vSlice)
			assert.Nil(t, err)
			fmt.Println(vSlice)

			var vMap map[string]*mapValue
			err = p.Get(&vMap)
			assert.Nil(t, err)
			fmt.Println(vMap)
		})
		assert.Nil(t, err)
	})
}

type circularA struct {
	b *circularB
}

func newCircularA(b *circularB) *circularA {
	return &circularA{b: b}
}

type circularB struct {
	A *circularA `autowire:",lazy"`
}

func newCircularB() *circularB {
	return new(circularB)
}

func TestLazy(t *testing.T) {
	for i := 0; i < 1; i++ {
		c := New()
		c.AllowCircularReferences()

		c.Provide(newCircularA)
		c.Provide(newCircularB)
		d := struct {
			b *circularB `autowire:""`
		}{}
		c.Object(&d)

		err := c.Properties().Refresh(conf.New())
		assert.Nil(t, err)
		err = c.Refresh()
		assert.Nil(t, err)
		assert.NotNil(t, d.b)
		assert.NotNil(t, d.b.A)
	}
}

type memory struct {
}

func (m *memory) OnInit(ctx Context) error {
	fmt.Println("memory.OnInit")
	return nil
}

func (m *memory) OnDestroy() {
	fmt.Println("memory.OnDestroy")
}

type table struct {
	_ *memory `autowire:""`
}

func (t *table) OnInit(ctx Context) error {
	fmt.Println("table.OnInit")
	return nil
}

func (t *table) OnDestroy() {
	fmt.Println("table.OnDestroy")
}

func TestDestroyDependence(t *testing.T) {
	c := New()
	c.Object(new(memory))
	c.Object(new(table)).Name("aaa")
	c.Object(new(table)).Name("bbb")
	err := c.Refresh()
	assert.Nil(t, err)
}

type TContextAware struct {
	ContextAware
}

func (c *TContextAware) Echo(str string) string {
	return c.GSContext.Prop("prefix") + " " + str + "!"
}

func TestContextAware(t *testing.T) {
	c := New()
	p := conf.New()
	p.Set("prefix", "hello")
	b := c.Object(new(TContextAware))

	err := c.Properties().Refresh(p)
	assert.Nil(t, err)
	err = c.Refresh()
	assert.Nil(t, err)
	a := b.Interface().(*TContextAware)
	assert.Equal(t, a.Echo("gopher"), "hello gopher!")
}

type TestLogApp struct {
	Logger     *Logger `logger:""` // primary logger
	StdLogger  *Logger `logger:":std"`
	ErrLogger  *Logger `logger:":err"`
	JsonLogger *Logger `logger:"${app.logger}"`
}

func TestLogLogger(t *testing.T) {

	var stdWriter bytes.Buffer
	var errWriter bytes.Buffer
	var jsonWriter bytes.Buffer

	var stdLogger = slog.New(slog.NewTextHandler(&stdWriter, nil))
	var errLogger = slog.New(slog.NewTextHandler(&errWriter, nil))
	var jsonLogger = slog.New(slog.NewJSONHandler(&jsonWriter, nil))

	// 也可以在OnProperty回调中注册
	SetLogger("std", stdLogger, true)
	SetLogger("err", errLogger)
	SetLogger("json", jsonLogger)

	c := New()
	p := conf.New()
	p.Set("app.logger", "json")
	err := c.Properties().Refresh(p)
	assert.Nil(t, err)

	b := c.Object(new(TestLogApp))
	err = c.Refresh()
	assert.Nil(t, err)
	logApp := b.Interface().(*TestLogApp)

	logApp.Logger.Info("primary info message")
	assert.Matches(t, stdWriter.String(), "primary info message")

	stdWriter.Reset()

	logApp.Logger.Info("std info message")
	assert.Matches(t, stdWriter.String(), "std info message")

	logApp.Logger.Info("err info message")
	assert.Matches(t, stdWriter.String(), "err info message")

	logApp.Logger.Info("json info message")
	assert.Matches(t, stdWriter.String(), "json info message")

}
