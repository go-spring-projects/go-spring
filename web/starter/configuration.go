/*
 * Copyright 2023 the original author or authors.
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

package starter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-spring-projects/go-spring/gs"
	"github.com/go-spring-projects/go-spring/gs/arg"
	"github.com/go-spring-projects/go-spring/gs/cond"
	"github.com/go-spring-projects/go-spring/web"
)

func init() {
	gs.Configuration(new(serverConfiguration)).
		On(cond.OnProperty("http.addr"))
}

type serverConfiguration struct {
	Logger *slog.Logger `logger:""`
	Server *web.Server  `autowire:""`
}

func (sc *serverConfiguration) OnAppStart(ctx context.Context) {
	sc.Logger.Info("starting http server", slog.String("addr", sc.Server.Addr()))

	go func() {
		if err := sc.Server.Run(); nil != err && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("failed to start http server `%s`: %w", sc.Server.Addr(), err))
		}
	}()
}

func (sc *serverConfiguration) OnAppStop(ctx context.Context) {
	sc.Logger.Info("stopping http server", slog.String("addr", sc.Server.Addr()))

	if err := sc.Server.Shutdown(ctx); nil != err {
		sc.Logger.Error("http server shutdown failed", slog.String("addr", sc.Server.Addr()), slog.Any("err", err))
	} else {
		sc.Logger.Info("http server shutdown successfully", slog.String("addr", sc.Server.Addr()))
	}
}

func (sc *serverConfiguration) NewServer() *gs.BeanDefinition {
	return gs.NewBean(web.NewServer, arg.Value(web.NewRouter()), "${http}").Primary()
}
