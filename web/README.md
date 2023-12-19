# web
Move to https://github.com/go-spring-projects/web

## Quick start

```yaml
# http server config
http:
  # listen address
  addr: ":8080"
```

```go
package main

import (
	"context"

	"go-spring.dev/spring/gs"
	_ "go-spring.dev/spring/web/starter"
	"go-spring.dev/web"
)

type App struct {
	Router web.Router `autowire:""`
}

func (app *App) OnInit(ctx context.Context) error {
	app.Router.Get("/greeting", app.Greeting)
	return nil
}

func (app *App) Greeting(ctx context.Context) string {
	return "greeting!!!"
}

func main() {
	gs.Object(new(App))

	if err := gs.Run(); nil != err {
		panic(err)
	}
}

```