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
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

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

// App 应用
type App struct {
	logger    *log.Logger
	container *container
	exitChan  chan struct{}

	Events  []AppEvent  `autowire:"*?"`
	Runners []AppRunner `autowire:"*?"`
}

// NewApp application 的构造函数
func NewApp() *App {
	return &App{
		container: New().(*container),
		exitChan:  make(chan struct{}),
	}
}

func (app *App) Run() error {
	app.Object(app)
	app.logger = log.GetLogger(utils.TypeName(app))

	// 响应控制台的 Ctrl+C 及 kill 命令。
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		app.Shutdown(fmt.Sprintf("signal %v", sig))
	}()

	if err := app.start(); err != nil {
		return err
	}

	<-app.exitChan

	app.container.Close()
	app.logger.Info("application exited")
	return nil
}

func (app *App) clear() {
	app.container.clear()
}

func (app *App) start() error {

	e := NewConfiguration(new(FileResourceLocator))

	if err := e.Load(app.container.initProperties); nil != err {
		return err
	}

	if showBanner, _ := strconv.ParseBool(app.container.initProperties.Get(SpringBannerVisible)); showBanner {
		app.printBanner(app.getBanner(app.container.initProperties))
	}

	// 初始化属性
	if err := app.container.p.Refresh(app.container.initProperties); nil != err {
		return err
	}

	// 执行依赖注入
	if err := app.container.refresh(false); err != nil {
		return err
	}

	// 执行命令行启动器
	for _, r := range app.Runners {
		r.Run(app.container)
	}

	// 通知应用启动事件
	for _, event := range app.Events {
		event.OnAppStart(app.container)
	}

	// 通知应用停止事件
	app.container.Go(func(ctx context.Context) {
		<-ctx.Done()
		for _, event := range app.Events {
			event.OnAppStop(context.Background())
		}
	})

	app.logger.Info("application started successfully")
	app.clear()
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

func (app *App) getBanner(p *conf.Properties) string {
	var maxChars = 0
	if lines := strings.Split(DefaultBanner, "\n"); len(lines) > 0 {
		for _, line := range lines {
			if lineChars := utf8.RuneCountInString(line); lineChars > maxChars {
				maxChars = lineChars
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(DefaultBanner)

	sb.WriteString("\n")
	if chars := utf8.RuneCountInString(Version); maxChars > chars {
		sb.WriteString(strings.Repeat(" ", maxChars-chars))
	}
	sb.WriteString(Version)

	sb.WriteString("\n")
	if chars := utf8.RuneCountInString(Website); maxChars > chars {
		sb.WriteString(strings.Repeat(" ", maxChars-chars))
	}
	sb.WriteString(Website)

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
