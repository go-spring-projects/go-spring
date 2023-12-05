package render

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestRedirectRenderer(t *testing.T) {
	req, err := http.NewRequest("GET", "/test-redirect", nil)
	assert.Nil(t, err)

	data1 := RedirectRenderer{
		Code:     http.StatusMovedPermanently,
		Request:  req,
		Location: "/new/location",
	}

	w := httptest.NewRecorder()
	err = data1.Render(w)
	assert.Nil(t, err)

	data2 := RedirectRenderer{
		Code:     http.StatusOK,
		Request:  req,
		Location: "/new/location",
	}

	w = httptest.NewRecorder()
	assert.Panic(t, func() {
		err := data2.Render(w)
		assert.Nil(t, err)
	}, "Cannot redirect with status code 200")

	data3 := RedirectRenderer{
		Code:     http.StatusCreated,
		Request:  req,
		Location: "/new/location",
	}

	w = httptest.NewRecorder()
	err = data3.Render(w)
	assert.Nil(t, err)
}
