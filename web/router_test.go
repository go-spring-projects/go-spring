package web

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func URLParam(r *http.Request, name string) string {
	if ctx := FromRouteContext(r.Context()); nil != ctx {
		v, _ := ctx.URLParams.Get(name)
		return v
	}
	return ""
}

func TestMuxBasic(t *testing.T) {
	var count uint64
	countermw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count++
			next.ServeHTTP(w, r)
		})
	}

	usermw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxKey{"user"}, "peter")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}

	exmw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKey{"ex"}, "a")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}

	logbuf := bytes.NewBufferString("")
	logmsg := "logmw test"
	logmw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logbuf.WriteString(logmsg)
			next.ServeHTTP(w, r)
		})
	}

	cxindex := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value(ctxKey{"user"}).(string)
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("hi %s", user)))
	}

	headPing := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ping", "1")
		w.WriteHeader(200)
	}

	createPing := func(w http.ResponseWriter, r *http.Request) {
		// create ....
		w.WriteHeader(201)
	}

	pingAll2 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ping all2"))
	}

	pingOne := func(w http.ResponseWriter, r *http.Request) {
		idParam := URLParam(r, "id")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("ping one id: %s", idParam)))
	}

	pingWoop := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("woop." + URLParam(r, "iidd")))
	}

	catchAll := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("catchall"))
	}

	m := NewRouter()
	m.Use(countermw)
	m.Use(usermw)
	m.Use(exmw)
	m.Use(logmw)
	m.Get("/", cxindex)
	m.Get("/ping/all2", pingAll2)

	m.Head("/ping", headPing)
	m.Post("/ping", createPing)
	m.Get("/ping/{id}", pingWoop)
	m.Get("/ping/{id}", pingOne) // expected to overwrite to pingOne handler
	m.Get("/ping/{iidd}/woop", pingWoop)
	m.HandleFunc("/admin/*", catchAll)
	// m.Post("/admin/*", catchAll)

	ts := httptest.NewServer(m)
	defer ts.Close()

	// GET /
	if _, body := testRequest(t, ts, "GET", "/", nil); body != "hi peter" {
		t.Fatalf(body)
	}
	tlogmsg, _ := logbuf.ReadString(0)
	if tlogmsg != logmsg {
		t.Error("expecting log message from middleware:", logmsg)
	}

	// GET /ping/all2
	if _, body := testRequest(t, ts, "GET", "/ping/all2", nil); body != "ping all2" {
		t.Fatalf(body)
	}

	// GET /ping/123
	if _, body := testRequest(t, ts, "GET", "/ping/123", nil); body != "ping one id: 123" {
		t.Fatalf(body)
	}

	// GET /ping/allan
	if _, body := testRequest(t, ts, "GET", "/ping/allan", nil); body != "ping one id: allan" {
		t.Fatalf(body)
	}

	// GET /ping/1/woop
	if _, body := testRequest(t, ts, "GET", "/ping/1/woop", nil); body != "woop.1" {
		t.Fatalf(body)
	}

	// HEAD /ping
	resp, err := http.Head(ts.URL + "/ping")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("head failed, should be 200")
	}
	if resp.Header.Get("X-Ping") == "" {
		t.Error("expecting X-Ping header")
	}

	// GET /admin/catch-this
	if _, body := testRequest(t, ts, "GET", "/admin/catch-thazzzzz", nil); body != "catchall" {
		t.Fatalf(body)
	}

	// POST /admin/catch-this
	resp, err = http.Post(ts.URL+"/admin/casdfsadfs", "text/plain", bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Error("POST failed, should be 200")
	}

	if string(body) != "catchall" {
		t.Error("expecting response body: 'catchall'")
	}

	// Custom http method DIE /ping/1/woop
	if resp, body := testRequest(t, ts, "DIE", "/ping/1/woop", nil); body != "405 method not allowed\n" || resp.StatusCode != 405 {
		t.Fatalf(fmt.Sprintf("expecting 405 status and empty body, got %d '%s'", resp.StatusCode, body))
	}
}

func TestMuxMounts(t *testing.T) {
	r := NewRouter()

	r.Get("/{hash}", func(w http.ResponseWriter, r *http.Request) {
		v := URLParam(r, "hash")
		w.Write([]byte(fmt.Sprintf("/%s", v)))
	})

	(func(r Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			v := URLParam(r, "hash")
			w.Write([]byte(fmt.Sprintf("/%s/share", v)))
		})
		r.Get("/{network}", func(w http.ResponseWriter, r *http.Request) {
			v := URLParam(r, "hash")
			n := URLParam(r, "network")
			w.Write([]byte(fmt.Sprintf("/%s/share/%s", v, n)))
		})
	})(r.Group("/{hash}/share"))

	m := NewRouter().(*routerGroup)
	m.Mount("/sharing", r)

	ts := httptest.NewServer(m)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/sharing/aBc", nil); body != "/aBc" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/sharing/aBc/share", nil); body != "/aBc/share" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/sharing/aBc/share/twitter", nil); body != "/aBc/share/twitter" {
		t.Fatalf(body)
	}
}

func TestMuxPlain(t *testing.T) {
	r := NewRouter()
	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bye"))
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("nothing here"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/hi", nil); body != "bye" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/nothing-here", nil); body != "nothing here" {
		t.Fatalf(body)
	}
}

func TestMuxEmptyRoutes(t *testing.T) {
	mux := NewRouter()

	apiRouter := NewRouter()
	// oops, we forgot to declare any route handlers

	mux.Handle("/api*", apiRouter)

	if _, body := testHandler(t, mux, "GET", "/", nil); body != "404 page not found\n" {
		t.Fatalf(body)
	}

	if _, body := testHandler(t, apiRouter, "GET", "/", nil); body != "404 page not found\n" {
		t.Fatalf(body)
	}
}

// Test a mux that routes a trailing slash, see also middleware/strip_test.go
// for an example of using a middleware to handle trailing slashes.
func TestMuxTrailingSlash(t *testing.T) {
	r := NewRouter().(*routerGroup)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("nothing here"))
	})

	subRoutes := NewRouter()
	indexHandler := func(w http.ResponseWriter, r *http.Request) {
		accountID := URLParam(r, "accountID")
		w.Write([]byte(accountID))
	}
	subRoutes.Get("/", indexHandler)

	r.Mount("/accounts/{accountID}", subRoutes)
	r.Get("/accounts/{accountID}/", indexHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/accounts/admin", nil); body != "admin" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/accounts/admin/", nil); body != "admin" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/nothing-here", nil); body != "nothing here" {
		t.Fatalf(body)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	r := NewRouter()

	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi, get"))
	})

	r.Head("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi, head"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("Registered Method", func(t *testing.T) {
		resp, _ := testRequest(t, ts, "GET", "/hi", nil)
		if resp.StatusCode != 200 {
			t.Fatal(resp.Status)
		}
		if resp.Header.Values("Allow") != nil {
			t.Fatal("allow should be empty when method is registered")
		}
	})

	t.Run("Unregistered Method", func(t *testing.T) {
		resp, _ := testRequest(t, ts, "POST", "/hi", nil)
		if resp.StatusCode != 405 {
			t.Fatal(resp.Status)
		}
	})
}

func TestMuxNestedMethodNotAllowed(t *testing.T) {
	r := NewRouter().(*routerGroup)
	r.Get("/root", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("root 405"))
	})

	sr1 := NewRouter()
	sr1.Get("/sub1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("sub1"))
	})
	sr1.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("sub1 405"))
	})

	sr2 := NewRouter()
	sr2.Get("/sub2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("sub2"))
	})

	pathVar := NewRouter()
	pathVar.Get("/{var}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pv"))
	})
	pathVar.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("pv 405"))
	})

	r.Mount("/prefix1", sr1)
	r.Mount("/prefix2", sr2)
	r.Mount("/pathVar", pathVar)

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/root", nil); body != "root" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "PUT", "/root", nil); body != "root 405" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/prefix1/sub1", nil); body != "sub1" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "PUT", "/prefix1/sub1", nil); body != "sub1 405" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/prefix2/sub2", nil); body != "sub2" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "PUT", "/prefix2/sub2", nil); body != "root 405" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/pathVar/myvar", nil); body != "pv" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "DELETE", "/pathVar/myvar", nil); body != "pv 405" {
		t.Fatalf(body)
	}
}

func TestMuxComplicatedNotFound(t *testing.T) {
	decorateRouter := func(r *routerGroup) {
		// Root router with groups
		r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("auth get"))
		})
		(func(r Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("public get"))
			})
		})(r.Group("/public"))

		// sub router with groups
		sub0 := NewRouter()
		(func(r Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("private get"))
			})
		})(sub0.Group("/resource"))
		r.Mount("/private", sub0)

		// sub router with groups
		sub1 := NewRouter()
		(func(r Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("private get"))
			})
		})(sub1.Group("/resource"))
	}

	testNotFound := func(t *testing.T, r *routerGroup) {
		ts := httptest.NewServer(r)
		defer ts.Close()

		// check that we didn't break correct routes
		if _, body := testRequest(t, ts, "GET", "/auth", nil); body != "auth get" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/public", nil); body != "public get" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/public/", nil); body != "public get" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/private/resource", nil); body != "private get" {
			t.Fatalf(body)
		}
		// check custom not-found on all levels
		if _, body := testRequest(t, ts, "GET", "/nope", nil); body != "custom not-found" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/public/nope", nil); body != "custom not-found" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/private/nope", nil); body != "custom not-found" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/private/resource/nope", nil); body != "custom not-found" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/private_mw/nope", nil); body != "custom not-found" {
			t.Fatalf(body)
		}
		if _, body := testRequest(t, ts, "GET", "/private_mw/resource/nope", nil); body != "custom not-found" {
			t.Fatalf(body)
		}
		// check custom not-found on trailing slash routes
		if _, body := testRequest(t, ts, "GET", "/auth/", nil); body != "custom not-found" {
			t.Fatalf(body)
		}
	}

	t.Run("pre", func(t *testing.T) {
		r := NewRouter().(*routerGroup)
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("custom not-found"))
		})
		decorateRouter(r)
		testNotFound(t, r)
	})

	t.Run("post", func(t *testing.T) {
		r := NewRouter().(*routerGroup)
		decorateRouter(r)
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("custom not-found"))
		})
		testNotFound(t, r)
	})
}

func TestMuxMiddlewareStack(t *testing.T) {
	var stdmwInit, stdmwHandler uint64
	stdmw := func(next http.Handler) http.Handler {
		stdmwInit++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stdmwHandler++
			next.ServeHTTP(w, r)
		})
	}
	_ = stdmw

	var ctxmwInit, ctxmwHandler uint64
	ctxmw := func(next http.Handler) http.Handler {
		ctxmwInit++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxmwHandler++
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxKey{"count.ctxmwHandler"}, ctxmwHandler)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}

	r := NewRouter()
	r.Use(stdmw)
	r.Use(ctxmw)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/ping" {
				w.Write([]byte("pong"))
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	var handlerCount uint64

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlerCount++
		ctx := r.Context()
		ctxmwHandlerCount := ctx.Value(ctxKey{"count.ctxmwHandler"}).(uint64)
		w.Write([]byte(fmt.Sprintf("inits:%d reqs:%d ctxValue:%d", ctxmwInit, handlerCount, ctxmwHandlerCount)))
	})

	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("wooot"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	testRequest(t, ts, "GET", "/", nil)
	testRequest(t, ts, "GET", "/", nil)
	var body string
	_, body = testRequest(t, ts, "GET", "/", nil)
	if body != "inits:1 reqs:3 ctxValue:3" {
		t.Fatalf("got: '%s'", body)
	}

	_, body = testRequest(t, ts, "GET", "/ping", nil)
	if body != "pong" {
		t.Fatalf("got: '%s'", body)
	}
}

func TestMuxSubroutesBasic(t *testing.T) {
	hIndex := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("index"))
	})
	hArticlesList := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("articles-list"))
	})
	hSearchArticles := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("search-articles"))
	})
	hGetArticle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("get-article:%s", URLParam(r, "id"))))
	})
	hSyncArticle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("sync-article:%s", URLParam(r, "id"))))
	})

	r := NewRouter()
	// var rr1, rr2 *Mux
	r.Get("/", hIndex)
	(func(r Router) {
		// rr1 = r.(*Mux)
		r.Get("/", hArticlesList)
		r.Get("/search", hSearchArticles)
		(func(r Router) {
			// rr2 = r.(*Mux)
			r.Get("/", hGetArticle)
			r.Get("/sync", hSyncArticle)
		})(r.Group("/{id}"))
	})(r.Group("/articles"))

	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")
	// debugPrintTree(0, 0, r.tree, 0)
	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")

	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")
	// debugPrintTree(0, 0, rr1.tree, 0)
	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")

	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")
	// debugPrintTree(0, 0, rr2.tree, 0)
	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")

	ts := httptest.NewServer(r)
	defer ts.Close()

	var body, expected string

	_, body = testRequest(t, ts, "GET", "/", nil)
	expected = "index"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/articles", nil)
	expected = "articles-list"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/articles/search", nil)
	expected = "search-articles"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/articles/123", nil)
	expected = "get-article:123"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/articles/123/sync", nil)
	expected = "sync-article:123"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
}

func TestMuxSubroutes(t *testing.T) {
	hHubView1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hub1"))
	})
	hHubView2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hub2"))
	})
	hHubView3 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hub3"))
	})
	hAccountView1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("account1"))
	})
	hAccountView2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("account2"))
	})

	r := NewRouter().(*routerGroup)
	r.Get("/hubs/{hubID}/view", hHubView1)
	r.Get("/hubs/{hubID}/view/*", hHubView2)

	sr := NewRouter().(*routerGroup)
	sr.Get("/", hHubView3)
	r.Mount("/hubs/{hubID}/users", sr)
	r.Get("/hubs/{hubID}/users/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hub3 override"))
	})

	sr3 := NewRouter()
	sr3.Get("/", hAccountView1)
	sr3.Get("/hi", hAccountView2)

	// var sr2 *Mux
	(func(r Router) {
		rg := r.(*routerGroup) // sr2
		// r.Get("/", hAccountView1)
		rg.Mount("/", sr3)
	})(r.Group("/accounts/{accountID}"))

	// This is the same as the r.Route() call mounted on sr2
	// sr2 := NewRouter()
	// sr2.Mount("/", sr3)
	// r.Mount("/accounts/{accountID}", sr2)

	ts := httptest.NewServer(r)
	defer ts.Close()

	var body, expected string

	_, body = testRequest(t, ts, "GET", "/hubs/123/view", nil)
	expected = "hub1"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/hubs/123/view/index.html", nil)
	expected = "hub2"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/hubs/123/users", nil)
	expected = "hub3"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/hubs/123/users/", nil)
	expected = "hub3 override"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/accounts/44", nil)
	expected = "account1"
	if body != expected {
		t.Fatalf("request:%s expected:%s got:%s", "GET /accounts/44", expected, body)
	}
	_, body = testRequest(t, ts, "GET", "/accounts/44/hi", nil)
	expected = "account2"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}

	// Test that we're building the routingPatterns properly
	router := r
	req, _ := http.NewRequest("GET", "/accounts/44/hi", nil)

	rctx := &RouteContext{}
	req = req.WithContext(context.WithValue(req.Context(), routeContextKey{}, rctx))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body = w.Body.String()
	expected = "account2"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}

	routePatterns := rctx.routePatterns
	if len(rctx.routePatterns) != 3 {
		t.Fatalf("expected 3 routing patterns, got:%d", len(rctx.routePatterns))
	}
	expected = "/accounts/{accountID}/*"
	if routePatterns[0] != expected {
		t.Fatalf("routePattern, expected:%s got:%s", expected, routePatterns[0])
	}
	expected = "/*"
	if routePatterns[1] != expected {
		t.Fatalf("routePattern, expected:%s got:%s", expected, routePatterns[1])
	}
	expected = "/hi"
	if routePatterns[2] != expected {
		t.Fatalf("routePattern, expected:%s got:%s", expected, routePatterns[2])
	}

}

func TestSingleHandler(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := URLParam(r, "name")
		w.Write([]byte("hi " + name))
	})

	r, _ := http.NewRequest("GET", "/", nil)
	rctx := &RouteContext{}
	r = r.WithContext(context.WithValue(r.Context(), routeContextKey{}, rctx))
	rctx.URLParams.Add("name", "joe")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	body := w.Body.String()
	expected := "hi joe"
	if body != expected {
		t.Fatalf("expected:%s got:%s", expected, body)
	}
}

// TODO: a Router wrapper test..
//
// type ACLMux struct {
// 	*Mux
// 	XX string
// }
//
// func NewACLMux() *ACLMux {
// 	return &ACLMux{Mux: NewRouter(), XX: "hihi"}
// }
//
// // TODO: this should be supported...
// func TestWoot(t *testing.T) {
// 	var r Router = NewRouter()
//
// 	var r2 Router = NewACLMux() //NewRouter()
// 	r2.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("hi"))
// 	})
//
// 	r.Mount("/", r2)
// }

func TestServeHTTPExistingContext(t *testing.T) {
	r := NewRouter()
	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		s, _ := r.Context().Value(ctxKey{"testCtx"}).(string)
		w.Write([]byte(s))
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		s, _ := r.Context().Value(ctxKey{"testCtx"}).(string)
		w.WriteHeader(404)
		w.Write([]byte(s))
	})

	testcases := []struct {
		Ctx            context.Context
		Method         string
		Path           string
		ExpectedBody   string
		ExpectedStatus int
	}{
		{
			Method:         "GET",
			Path:           "/hi",
			Ctx:            context.WithValue(context.Background(), ctxKey{"testCtx"}, "hi ctx"),
			ExpectedStatus: 200,
			ExpectedBody:   "hi ctx",
		},
		{
			Method:         "GET",
			Path:           "/hello",
			Ctx:            context.WithValue(context.Background(), ctxKey{"testCtx"}, "nothing here ctx"),
			ExpectedStatus: 404,
			ExpectedBody:   "nothing here ctx",
		},
	}

	for _, tc := range testcases {
		resp := httptest.NewRecorder()
		req, err := http.NewRequest(tc.Method, tc.Path, nil)
		if err != nil {
			t.Fatalf("%v", err)
		}
		req = req.WithContext(tc.Ctx)
		r.ServeHTTP(resp, req)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if resp.Code != tc.ExpectedStatus {
			t.Fatalf("%v != %v", tc.ExpectedStatus, resp.Code)
		}
		if string(b) != tc.ExpectedBody {
			t.Fatalf("%s != %s", tc.ExpectedBody, b)
		}
	}
}

func TestMiddlewarePanicOnLateUse(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello\n"))
	}

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	defer func() {
		if recover() == nil {
			t.Error("expected panic()")
		}
	}()

	r := NewRouter()
	r.Get("/", handler)
	r.Use(mw) // Too late to apply middleware, we're expecting panic().
}

func TestMountingExistingPath(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}

	defer func() {
		if recover() == nil {
			t.Error("expected panic()")
		}
	}()

	r := NewRouter().(*routerGroup)
	r.Get("/", handler)
	r.Mount("/hi", http.HandlerFunc(handler))
	r.Mount("/hi", http.HandlerFunc(handler))
}

func TestMountingSimilarPattern(t *testing.T) {
	r := NewRouter().(*routerGroup)
	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bye"))
	})

	r2 := NewRouter()
	r2.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("foobar"))
	})

	r3 := NewRouter()
	r3.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("foo"))
	})

	r.Mount("/foobar", r2)
	r.Mount("/foo", r3)

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/hi", nil); body != "bye" {
		t.Fatalf(body)
	}
}

func TestMuxEmptyParams(t *testing.T) {
	r := NewRouter()
	r.Get(`/users/{x}/{y}/{z}`, func(w http.ResponseWriter, r *http.Request) {
		x := URLParam(r, "x")
		y := URLParam(r, "y")
		z := URLParam(r, "z")
		w.Write([]byte(fmt.Sprintf("%s-%s-%s", x, y, z)))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/users/a/b/c", nil); body != "a-b-c" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/users///c", nil); body != "--c" {
		t.Fatalf(body)
	}
}

func TestMuxMissingParams(t *testing.T) {
	r := NewRouter()
	r.Get(`/user/{userId:\d+}`, func(w http.ResponseWriter, r *http.Request) {
		userID := URLParam(r, "userId")
		w.Write([]byte(fmt.Sprintf("userId = '%s'", userID)))
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("nothing here"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/user/123", nil); body != "userId = '123'" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/user/", nil); body != "nothing here" {
		t.Fatalf(body)
	}
}

func TestMuxWildcardRoute(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}

	defer func() {
		if recover() == nil {
			t.Error("expected panic()")
		}
	}()

	r := NewRouter()
	r.Get("/*/wildcard/must/be/at/end", handler)
}

func TestMuxWildcardRouteCheckTwo(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}

	defer func() {
		if recover() == nil {
			t.Error("expected panic()")
		}
	}()

	r := NewRouter()
	r.Get("/*/wildcard/{must}/be/at/end", handler)
}

func TestMuxRegexp(t *testing.T) {
	r := NewRouter()
	r.Group("/{param:[0-9]*}/test", func(r Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(fmt.Sprintf("Hi: %s", URLParam(r, "param"))))
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "//test", nil); body != "Hi: " {
		t.Fatalf(body)
	}
}

func TestMuxRegexp2(t *testing.T) {
	r := NewRouter()
	r.Get("/foo-{suffix:[a-z]{2,3}}.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(URLParam(r, "suffix")))
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/foo-.json", nil); body != "" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/foo-abc.json", nil); body != "abc" {
		t.Fatalf(body)
	}
}

func TestMuxRegexp3(t *testing.T) {
	r := NewRouter()
	r.Get("/one/{firstId:[a-z0-9-]+}/{secondId:[a-z]+}/first", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("first"))
	})
	r.Get("/one/{firstId:[a-z0-9-_]+}/{secondId:[0-9]+}/second", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("second"))
	})
	r.Delete("/one/{firstId:[a-z0-9-_]+}/{secondId:[0-9]+}/second", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("third"))
	})

	(func(r Router) {
		r.Get("/{dns:[a-z-0-9_]+}", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("_"))
		})
		r.Get("/{dns:[a-z-0-9_]+}/info", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("_"))
		})
		r.Delete("/{id:[0-9]+}", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("forth"))
		})
	})(r.Group("/one"))

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/one/hello/peter/first", nil); body != "first" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/one/hithere/123/second", nil); body != "second" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "DELETE", "/one/hithere/123/second", nil); body != "third" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "DELETE", "/one/123", nil); body != "forth" {
		t.Fatalf(body)
	}
}

func TestMuxSubrouterWildcardParam(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "param:%v *:%v", URLParam(r, "param"), URLParam(r, "*"))
	})

	r := NewRouter()

	r.Get("/bare/{param}", h)
	r.Get("/bare/{param}/*", h)

	(func(r Router) {
		r.Get("/{param}", h)
		r.Get("/{param}/*", h)
	})(r.Group("/case0"))

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/bare/hi", nil); body != "param:hi *:" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/bare/hi/yes", nil); body != "param:hi *:yes" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/case0/hi", nil); body != "param:hi *:" {
		t.Fatalf(body)
	}
	if _, body := testRequest(t, ts, "GET", "/case0/hi/yes", nil); body != "param:hi *:yes" {
		t.Fatalf(body)
	}
}

func TestMuxContextIsThreadSafe(t *testing.T) {
	router := NewRouter()
	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Millisecond)
		defer cancel()

		<-ctx.Done()
	})

	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				w := httptest.NewRecorder()
				r, err := http.NewRequest("GET", "/ok", nil)
				if err != nil {
					t.Error(err)
					return
				}

				ctx, cancel := context.WithCancel(r.Context())
				r = r.WithContext(ctx)

				go func() {
					cancel()
				}()
				router.ServeHTTP(w, r)
			}
		}()
	}
	wg.Wait()
}

func TestEscapedURLParams(t *testing.T) {
	m := NewRouter()
	m.Get("/api/{identifier}/{region}/{size}/{rotation}/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		rctx := FromRouteContext(r.Context())
		if rctx == nil {
			t.Error("no context")
			return
		}
		identifier := URLParam(r, "identifier")
		if identifier != "http:%2f%2fexample.com%2fimage.png" {
			t.Errorf("identifier path parameter incorrect %s", identifier)
			return
		}
		region := URLParam(r, "region")
		if region != "full" {
			t.Errorf("region path parameter incorrect %s", region)
			return
		}
		size := URLParam(r, "size")
		if size != "max" {
			t.Errorf("size path parameter incorrect %s", size)
			return
		}
		rotation := URLParam(r, "rotation")
		if rotation != "0" {
			t.Errorf("rotation path parameter incorrect %s", rotation)
			return
		}
		w.Write([]byte("success"))
	})

	ts := httptest.NewServer(m)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/api/http:%2f%2fexample.com%2fimage.png/full/max/0/color.png", nil); body != "success" {
		t.Fatalf(body)
	}
}

func TestMuxMatch(t *testing.T) {
	r := NewRouter()
	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "yes")
		w.Write([]byte("bye"))
	})
	(func(r Router) {
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := URLParam(r, "id")
			w.Header().Set("X-Article", id)
			w.Write([]byte("article:" + id))
		})
	})(r.Group("/articles"))
	(func(r Router) {
		r.Head("/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-User", "-")
			w.Write([]byte("user"))
		})
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := URLParam(r, "id")
			w.Header().Set("X-User", id)
			w.Write([]byte("user:" + id))
		})
	})(r.Group("/users"))

	tctx := &RouteContext{}

	tctx.Reset()
	if r.(Routes).Match(tctx, "GET", "/users/1") == false {
		t.Fatal("expecting to find match for route:", "GET", "/users/1")
	}

	tctx.Reset()
	if r.(Routes).Match(tctx, "HEAD", "/articles/10") == true {
		t.Fatal("not expecting to find match for route:", "HEAD", "/articles/10")
	}
}

func TestServerBaseContext(t *testing.T) {
	r := NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		baseYes := r.Context().Value(ctxKey{"base"}).(string)
		if _, ok := r.Context().Value(http.ServerContextKey).(*http.Server); !ok {
			panic("missing server context")
		}
		if _, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); !ok {
			panic("missing local addr context")
		}
		w.Write([]byte(baseYes))
	})

	// Setup http Server with a base context
	ctx := context.WithValue(context.Background(), ctxKey{"base"}, "yes")
	ts := httptest.NewUnstartedServer(r)
	ts.Config.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}
	ts.Start()

	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/", nil); body != "yes" {
		t.Fatalf(body)
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}

func testHandler(t *testing.T, h http.Handler, method, path string, body io.Reader) (*http.Response, string) {
	r, _ := http.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Result(), w.Body.String()
}

type ctxKey struct {
	name string
}

func (k ctxKey) String() string {
	return "context value " + k.name
}

func BenchmarkMux(b *testing.B) {
	h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h3 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h4 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h5 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h6 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mx := NewRouter()
	mx.Get("/", h1)
	mx.Get("/hi", h2)
	mx.Post("/hi-post", h2) // used to benchmark 405 responses
	mx.Get("/sup/{id}/and/{this}", h3)
	mx.Get("/sup/{id}/{bar:foo}/{this}", h3)

	mx.Group("/sharing/{x}/{hash}", func(mx Router) {
		mx.Get("/", h4)          // subrouter-1
		mx.Get("/{network}", h5) // subrouter-1
		mx.Get("/twitter", h5)
		mx.Group("/direct", func(mx Router) {
			mx.Get("/", h6) // subrouter-2
			mx.Get("/download", h6)
		})
	})

	routes := []string{
		"/",
		"/hi",
		"/hi-post",
		"/sup/123/and/this",
		"/sup/123/foo/this",
		"/sharing/z/aBc",                 // subrouter-1
		"/sharing/z/aBc/twitter",         // subrouter-1
		"/sharing/z/aBc/direct",          // subrouter-2
		"/sharing/z/aBc/direct/download", // subrouter-2
	}

	for _, path := range routes {
		b.Run("route:"+path, func(b *testing.B) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", path, nil)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				mx.ServeHTTP(w, r)
			}
		})
	}
}
