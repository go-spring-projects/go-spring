package render

import (
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestHTMLRenderer(t *testing.T) {

	w := httptest.NewRecorder()
	templ := template.Must(template.New("t").Parse(`Hello {{.name}}`))

	htmlRender := HTMLRenderer{Template: templ, Name: "t", Data: map[string]interface{}{"name": "asdklajhdasdd"}}
	err := htmlRender.Render(w)

	assert.Nil(t, err)
	assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8")
	assert.Equal(t, w.Body.String(), "Hello asdklajhdasdd")
}
