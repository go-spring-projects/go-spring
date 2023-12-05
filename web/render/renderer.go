package render

import "net/http"

// Renderer writes data with custom ContentType and headers.
type Renderer interface {
	Render(writer http.ResponseWriter) error
}
