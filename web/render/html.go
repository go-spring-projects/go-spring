package render

import (
	"html/template"
	"net/http"
)

type HTMLRenderer struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

func (h HTMLRenderer) Render(writer http.ResponseWriter) error {
	if header := writer.Header(); len(header.Get("Content-Type")) == 0 {
		header.Set("Content-Type", "text/html; charset=utf-8")
	}
	if len(h.Name) > 0 {
		return h.Template.ExecuteTemplate(writer, h.Name, h.Data)
	}
	return h.Template.Execute(writer, h.Data)
}
