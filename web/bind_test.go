package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-spring.dev/spring/internal/utils/assert"
)

func TestBindWithoutParams(t *testing.T) {

	var handler = func(ctx context.Context) string {
		webCtx := FromContext(ctx)
		assert.NotNil(t, webCtx)
		return "0987654321"
	}

	request := httptest.NewRequest(http.MethodGet, "/get", strings.NewReader("{}"))
	response := httptest.NewRecorder()
	Bind(handler, RendererFunc(defaultJsonRender))(response, request)
	assert.Equal(t, response.Body.String(), "{\"code\":0,\"data\":\"0987654321\"}\n")
}

func TestBindWithParams(t *testing.T) {
	var handler = func(ctx context.Context, req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}) string {
		webCtx := FromContext(ctx)
		assert.NotNil(t, webCtx)
		assert.Equal(t, req.Username, "aaa")
		assert.Equal(t, req.Password, "88888888")
		return "success"
	}

	request := httptest.NewRequest(http.MethodPost, "/post", strings.NewReader(`{"username": "aaa", "password": "88888888"}`))
	request.Header.Add("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Bind(handler, RendererFunc(defaultJsonRender))(response, request)
	assert.Equal(t, response.Body.String(), "{\"code\":0,\"data\":\"success\"}\n")
}

func TestBindWithParamsAndError(t *testing.T) {
	var handler = func(ctx context.Context, req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}) (string, error) {
		webCtx := FromContext(ctx)
		assert.NotNil(t, webCtx)
		assert.Equal(t, req.Username, "aaa")
		assert.Equal(t, req.Password, "88888888")
		return "requestid: 9999999", Error(403, "user locked")
	}

	request := httptest.NewRequest(http.MethodPost, "/post", strings.NewReader(`{"username": "aaa", "password": "88888888"}`))
	request.Header.Add("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Bind(handler, RendererFunc(defaultJsonRender))(response, request)
	assert.Equal(t, response.Body.String(), "{\"code\":403,\"message\":\"user locked\",\"data\":\"requestid: 9999999\"}\n")
}
