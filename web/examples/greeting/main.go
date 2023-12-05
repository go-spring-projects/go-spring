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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"mime/multipart"
	"time"

	"github.com/go-spring-projects/go-spring/gs"
	"github.com/go-spring-projects/go-spring/web"
	_ "github.com/go-spring-projects/go-spring/web/starter"
)

type Greeting struct {
	Logger *slog.Logger `logger:""`
	Server *web.Server  `autowire:""`
}

func (g *Greeting) OnInit(ctx context.Context) error {
	g.Server.Bind("/greeting", g.Greeting)
	g.Server.Bind("/health", g.Health)
	g.Server.Bind("/user/register/{username}/{password}", g.Register)
	return nil
}

func (g *Greeting) Greeting(ctx context.Context) string {
	web.FromContext(ctx).
		SetCookie("token", "1234567890", 500, "", "", false, false)
	return "greeting!!!"
}

func (g *Greeting) Health(ctx context.Context) (string, error) {
	if 0 == rand.Int()%2 {
		return "", fmt.Errorf("health check failed")
	}
	return time.Now().String(), nil
}

func (g *Greeting) Register(
	ctx context.Context,
	req struct {
		Username  string                `path:"username"`     // 用户名
		Password  string                `path:"password"`     // 密码
		HeadImg   *multipart.FileHeader `form:"headImg"`      // 上传头像
		Captcha   string                `form:"captcha"`      // 验证码
		UserAgent string                `header:"User-Agent"` // 用户代理
		Ad        string                `query:"ad"`          // 推广ID
		Token     string                `cookie:"token"`      // cookie参数
	},
) string {
	g.Logger.Info("register user",
		slog.String("username", req.Username),
		slog.String("password", req.Password),
		slog.String("userAgent", req.UserAgent),
		slog.String("headImg", req.HeadImg.Filename),
		slog.String("captcha", req.Captcha),
		slog.String("userAgent", req.UserAgent),
		slog.String("ad", req.Ad),
		slog.String("token", req.Token),
	)
	return "ok"
}

func main() {
	gs.Object(new(Greeting))

	if err := gs.Run(); nil != err {
		panic(err)
	}
}
