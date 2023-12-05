package render

import (
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestJSONRenderer(t *testing.T) {
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}

	w := httptest.NewRecorder()

	err := JsonRenderer{Data: data}.Render(w)
	assert.Nil(t, err)

	assert.Equal(t, w.Header().Get("Content-Type"), "application/json; charset=utf-8")
	assert.Equal(t, w.Body.String(), "{\"foo\":\"bar\",\"html\":\"\\u003cb\\u003e\"}\n")
}
