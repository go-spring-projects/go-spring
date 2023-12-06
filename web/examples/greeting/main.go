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
	"net/http"
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
	g.Server.Bind("/user/password", g.UpdatePassword)

	g.Server.Use(func(handler http.Handler) http.Handler {

		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

			start := time.Now()
			handler.ServeHTTP(writer, request)
			g.Logger.Info("http handle cost",
				slog.String("path", request.URL.Path), slog.Duration("cost", time.Since(start)))
		})
	})

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
		Username  string                `path:"username" expr:"len($)>6 && len($)<20"` // username
		Password  string                `path:"password" expr:"len($)>6 && len($)<20"` // password
		HeadImg   *multipart.FileHeader `form:"headImg"`                               // upload head image
		Captcha   string                `form:"captcha" expr:"len($)==4"`              // captcha
		UserAgent string                `header:"User-Agent"`                          // user agent
		Ad        string                `query:"ad"`                                   // AD
		Token     string                `cookie:"token"`                               // token
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

func (g *Greeting) UpdatePassword(
	ctx context.Context,
	req struct {
		Password string `json:"password" expr:"len($) > 6 && len($) < 20"` // password
		Token    string `cookie:"token"`                                   // token
	},
) string {
	g.Logger.Info("change password", slog.String("password", req.Password))
	return "ok"
}

func main() {
	gs.Object(new(Greeting))

	if err := gs.Run(); nil != err {
		panic(err)
	}
}
