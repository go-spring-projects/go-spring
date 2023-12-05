package render

import (
	"encoding/json"
	"net/http"
)

type JsonRenderer struct {
	Prefix string
	Indent string
	Data   interface{}
}

func (j JsonRenderer) Render(writer http.ResponseWriter) error {
	if header := writer.Header(); len(header.Get("Content-Type")) == 0 {
		header.Set("Content-Type", "application/json; charset=utf-8")
	}
	encoder := json.NewEncoder(writer)
	if len(j.Prefix) > 0 || len(j.Indent) > 0 {
		encoder.SetIndent(j.Prefix, j.Indent)
	}
	return encoder.Encode(j.Data)
}
