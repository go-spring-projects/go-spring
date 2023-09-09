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
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/limpo1989/go-spring/conf"
	"github.com/limpo1989/go-spring/gs/arg"
	"github.com/limpo1989/go-spring/log"
	"github.com/limpo1989/go-spring/utils"
)

// AppRunner 命令行启动器接口
type AppRunner interface {
	Run(ctx Context)
}

// AppEvent 应用运行过程中的事件
type AppEvent interface {
	OnAppStart(ctx Context)        // 应用启动的事件
	OnAppStop(ctx context.Context) // 应用停止的事件
}

// App 应用
type App struct {
	logger    *log.Logger
	container *container
	exitChan  chan struct{}
}

// NewApp application 的构造函数
func NewApp() *App {
	return &App{
		container: New().(*container),
		exitChan:  make(chan struct{}),
	}
}

func (app *App) Run(resourceLocator ...ResourceLocator) error {
	app.logger = log.GetLogger(utils.TypeName(app))

	// 响应控制台的 Ctrl+C 及 kill 命令。
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		app.Shutdown(fmt.Sprintf("signal %v", sig))
	}()

	var locator ResourceLocator = new(FileResourceLocator)
	if len(resourceLocator) > 0 && resourceLocator[0] != nil {
		locator = resourceLocator[0]
	}

	return app.run(locator)
}

func (app *App) run(resourceLocator ResourceLocator) error {

	e := NewConfiguration(resourceLocator)

	if err := e.Load(app.container.props); nil != err {
		return err
	}

	if showBanner, _ := strconv.ParseBool(app.container.props.Get("spring.config.banner", conf.Def("true"))); showBanner {
		app.printBanner(app.getBanner(app.container.props))
	}

	// 初始化属性
	if err := app.container.p.Refresh(app.container.props); nil != err {
		return err
	}

	// 执行依赖注入
	if err := app.container.refresh(false); err != nil {
		return err
	}

	//  OnInit
	//  Run
	//  OnAppStart
	//  ---wait-signal---
	//  Stop GoRoutine
	//  OnAppStop
	//  OnDestroy

	app.onAppRun(app.container)

	app.onAppStart(app.container)

	app.container.clear()
	app.logger.Info("application started successfully")

	// 等待应用停止信号
	<-app.exitChan

	// 停止所有受GS管理的协程
	app.container.Cancel()

	// 结束应用
	app.onAppStop(context.Background())

	// 执行析构函数
	app.container.Close()

	app.logger.Info("application exited")

	return nil
}

func (app *App) onAppRun(ctx Context) {
	for _, bean := range app.container.Dependencies(true) {
		x := bean.Value().Interface()

		if ar, ok := x.(AppRunner); ok {
			ar.Run(ctx)
		}
	}
}

func (app *App) onAppStart(ctx Context) {
	for _, bean := range app.container.Dependencies(true) {
		x := bean.Value().Interface()

		if ae, ok := x.(AppEvent); ok {
			ae.OnAppStart(ctx)
		}
	}
}

func (app *App) onAppStop(ctx context.Context) {
	for _, bean := range app.container.Dependencies(false) {
		x := bean.Value().Interface()

		if ae, ok := x.(AppEvent); ok {
			ae.OnAppStop(ctx)
		}
	}
}

const DefaultBanner = `
  ______  _____      _______  _____   ______ _____ __   _  ______
 |  ____ |     | ___ |______ |_____] |_____/   |   | \  | |  ____
 |_____| |_____|     ______| |       |    \_ __|__ |  \_| |_____|
`

func (app *App) getBanner(p *conf.Properties) string {
	var maxPadding = 0
	if lines := strings.Split(DefaultBanner, "\n"); len(lines) > 0 {
		for _, line := range lines {
			if lineChars := utf8.RuneCountInString(line); lineChars > maxPadding {
				maxPadding = lineChars
			}
		}
	}

	var splitter = strings.Repeat("-", maxPadding)
	var appRuntime = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	var sb strings.Builder
	sb.WriteString(DefaultBanner)

	for _, info := range []string{Website, Version, appRuntime} {
		if len(info) > 0 {
			sb.WriteString("\n")
			if chars := utf8.RuneCountInString(info); maxPadding > chars {
				sb.WriteString(strings.Repeat(" ", maxPadding-chars))
			}
			sb.WriteString(info)
		}
	}
	sb.WriteString("\n")
	sb.WriteString(splitter)
	return sb.String()
}

// printBanner 打印 banner 到控制台
func (app *App) printBanner(banner string) {
	if banner[0] != '\n' {
		fmt.Println()
	}
	fmt.Println(banner)
	fmt.Println()
}

// Shutdown 关闭执行器
func (app *App) Shutdown(msg ...string) {
	app.logger.Sugar().Infof("program will exit %s", strings.Join(msg, " "))
	select {
	case <-app.exitChan:
		// chan 已关闭，无需再次关闭。
	default:
		close(app.exitChan)
	}
}

// OnProperty 当 key 对应的属性值准备好后发送一个通知。
func (app *App) OnProperty(key string, fn interface{}) {
	app.container.OnProperty(key, fn)
}

// Property 参考 Container.Property 的解释。
func (app *App) Property(key string, value interface{}) {
	app.container.Property(key, value)
}

// Accept 参考 Container.Accept 的解释。
func (app *App) Accept(b *BeanDefinition) *BeanDefinition {
	return app.container.Accept(b)
}

// Object 参考 Container.Object 的解释。
func (app *App) Object(i interface{}) *BeanDefinition {
	return app.container.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func (app *App) Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition {
	return app.container.Accept(NewBean(ctor, args...))
}

// AllowCircularReferences 参考 Container.AllowCircularReferences 的解释。
func (app *App) AllowCircularReferences() {
	app.container.AllowCircularReferences()
}
