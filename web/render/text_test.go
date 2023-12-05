package render

import (
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestTextRenderer(t *testing.T) {
	w := httptest.NewRecorder()

	err := (TextRenderer{
		Format: "hello %s %d",
		Args:   []any{"bob", 2},
	}).Render(w)

	assert.Nil(t, err)
	assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
	assert.Equal(t, w.Body.String(), "hello bob 2")
}
