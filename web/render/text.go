package render

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type TextRenderer struct {
	Format string
	Args   []interface{}
}

func (t TextRenderer) Render(writer http.ResponseWriter) error {
	if header := writer.Header(); len(header.Get("Content-Type")) == 0 {
		header.Set("Content-Type", "text/plain; charset=utf-8")
	}
	_, err := io.Copy(writer, strings.NewReader(fmt.Sprintf(t.Format, t.Args...)))
	return err
}
