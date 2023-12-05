package render

import (
	"encoding/xml"
	"net/http"
)

type XmlRenderer struct {
	Prefix string
	Indent string
	Data   interface{}
}

func (x XmlRenderer) Render(writer http.ResponseWriter) error {
	if header := writer.Header(); len(header.Get("Content-Type")) == 0 {
		header.Set("Content-Type", "application/xml; charset=utf-8")
	}

	encoder := xml.NewEncoder(writer)
	if len(x.Prefix) > 0 || len(x.Indent) > 0 {
		encoder.Indent(x.Prefix, x.Indent)
	}
	return encoder.Encode(x.Data)
}
