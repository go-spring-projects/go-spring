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
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"github.com/limpo1989/go-spring/conf"
	"github.com/limpo1989/go-spring/gs/arg"
	"github.com/limpo1989/go-spring/log"
	"github.com/limpo1989/go-spring/utils"
)

// SpringBannerVisible 是否显示 banner。
const SpringBannerVisible = "spring.banner.visible"

// AppRunner 命令行启动器接口
type AppRunner interface {
	Run(ctx Context)
}

// AppEvent 应用运行过程中的事件
type AppEvent interface {
	OnAppStart(ctx Context)        // 应用启动的事件
	OnAppStop(ctx context.Context) // 应用停止的事件
}

type tempApp struct {
	banner string
}

// App 应用
type App struct {
	*tempApp

	logger *log.Logger

	c *container
	b *bootstrap

	exitChan chan struct{}

	Events  []AppEvent  `autowire:"${application-event.collection:=*?}"`
	Runners []AppRunner `autowire:"${command-line-runner.collection:=*?}"`
}

// NewApp application 的构造函数
func NewApp() *App {
	return &App{
		c:        New().(*container),
		tempApp:  &tempApp{},
		exitChan: make(chan struct{}),
	}
}

// Banner 自定义 banner 字符串。
func (app *App) Banner(banner string) {
	app.banner = banner
}

func (app *App) Run() error {
	app.Object(app)
	app.logger = log.GetLogger(utils.TypeName(app))

	// 响应控制台的 Ctrl+C 及 kill 命令。
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		app.ShutDown(fmt.Sprintf("signal %v", sig))
	}()

	if err := app.start(); err != nil {
		return err
	}

	<-app.exitChan

	if app.b != nil {
		app.b.c.Close()
	}

	app.c.Close()
	app.logger.Info("application exited")
	return nil
}

func (app *App) clear() {
	app.c.clear()
	if app.b != nil {
		app.b.clear()
	}
	app.tempApp = nil
}

func (app *App) start() error {

	e := &configuration{
		p:               conf.New(),
		resourceLocator: new(defaultResourceLocator),
	}

	if err := e.prepare(); err != nil {
		return err
	}

	showBanner, _ := strconv.ParseBool(e.p.Get(SpringBannerVisible))
	if showBanner {
		app.printBanner(app.getBanner(e))
	}

	if app.b != nil {
		if err := app.b.start(e); err != nil {
			return err
		}
	}

	if err := app.loadProperties(e); err != nil {
		return err
	}

	// 保存从环境变量和命令行解析的属性
	for _, k := range e.p.Keys() {
		app.c.initProperties.Set(k, e.p.Get(k))
	}

	// 初始化属性
	if err := app.c.p.Refresh(app.c.initProperties); nil != err {
		return err
	}

	if err := app.c.refresh(false); err != nil {
		return err
	}

	// 执行命令行启动器
	for _, r := range app.Runners {
		r.Run(app.c)
	}

	// 通知应用启动事件
	for _, event := range app.Events {
		event.OnAppStart(app.c)
	}

	app.clear()

	// 通知应用停止事件
	app.c.Go(func(ctx context.Context) {
		<-ctx.Done()
		for _, event := range app.Events {
			event.OnAppStop(context.Background())
		}
	})

	app.logger.Info("application started successfully")
	return nil
}

const DefaultBanner = `
                                              (_)              
  __ _    ___             ___   _ __    _ __   _   _ __     __ _ 
 / _' |  / _ \   ______  / __| | '_ \  | '__| | | | '_ \   / _' |
| (_| | | (_) | |______| \__ \ | |_) | | |    | | | | | | | (_| |
 \__, |  \___/           |___/ | .__/  |_|    |_| |_| |_|  \__, |
  __/ |                        | |                          __/ |
 |___/                         |_|                         |___/ 
`

func (app *App) getBanner(e *configuration) string {
	if app.banner != "" {
		return app.banner
	}
	resources, err := e.resourceLocator.Locate("banner.txt")
	if err != nil {
		return ""
	}
	banner := DefaultBanner
	for _, resource := range resources {
		if b, _ := ioutil.ReadAll(resource); b != nil {
			banner = string(b)
		}
	}
	return banner
}

// printBanner 打印 banner 到控制台
func (app *App) printBanner(banner string) {

	if banner[0] != '\n' {
		fmt.Println()
	}

	maxLength := 0
	for _, s := range strings.Split(banner, "\n") {
		fmt.Printf("\x1b[36m%s\x1b[0m\n", s) // CYAN
		if len(s) > maxLength {
			maxLength = len(s)
		}
	}

	if banner[len(banner)-1] != '\n' {
		fmt.Println()
	}

	var padding []byte
	if n := (maxLength - len(Version)) / 2; n > 0 {
		padding = make([]byte, n)
		for i := range padding {
			padding[i] = ' '
		}
	}
	fmt.Println(string(padding) + Version + "\n")
}

func (app *App) loadProperties(e *configuration) error {
	var resources []Resource

	for _, ext := range e.ConfigExtensions {
		sources, err := app.loadResource(e, "application"+ext)
		if err != nil {
			return err
		}
		resources = append(resources, sources...)
	}

	for _, profile := range e.ActiveProfiles {
		for _, ext := range e.ConfigExtensions {
			sources, err := app.loadResource(e, "application-"+profile+ext)
			if err != nil {
				return err
			}
			resources = append(resources, sources...)
		}
	}

	for _, resource := range resources {
		b, err := ioutil.ReadAll(resource)
		if err != nil {
			return err
		}
		p, err := conf.Bytes(b, filepath.Ext(resource.Name()))
		if err != nil {
			return err
		}
		for _, key := range p.Keys() {
			app.c.initProperties.Set(key, p.Get(key))
		}
	}

	return nil
}

func (app *App) loadResource(e *configuration, filename string) ([]Resource, error) {

	var locators []ResourceLocator
	locators = append(locators, e.resourceLocator)
	if app.b != nil {
		locators = append(locators, app.b.resourceLocators...)
	}

	var resources []Resource
	for _, locator := range locators {
		sources, err := locator.Locate(filename)
		if err != nil {
			return nil, err
		}
		resources = append(resources, sources...)
	}
	return resources, nil
}

// ShutDown 关闭执行器
func (app *App) ShutDown(msg ...string) {
	app.logger.Sugar().Infof("program will exit %s", strings.Join(msg, " "))
	select {
	case <-app.exitChan:
		// chan 已关闭，无需再次关闭。
	default:
		close(app.exitChan)
	}
}

// Bootstrap 返回 *bootstrap 对象。
func (app *App) Bootstrap() *bootstrap {
	if app.b == nil {
		app.b = newBootstrap()
	}
	return app.b
}

// OnProperty 当 key 对应的属性值准备好后发送一个通知。
func (app *App) OnProperty(key string, fn interface{}) {
	app.c.OnProperty(key, fn)
}

// Property 参考 Container.Property 的解释。
func (app *App) Property(key string, value interface{}) {
	app.c.Property(key, value)
}

// Accept 参考 Container.Accept 的解释。
func (app *App) Accept(b *BeanDefinition) *BeanDefinition {
	return app.c.Accept(b)
}

// Object 参考 Container.Object 的解释。
func (app *App) Object(i interface{}) *BeanDefinition {
	return app.c.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func (app *App) Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition {
	return app.c.Accept(NewBean(ctor, args...))
}
