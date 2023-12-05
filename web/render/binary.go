package render

import (
	"net/http"
)

type BinaryRenderer struct {
	ContentType string
	Data        []byte
}

func (b BinaryRenderer) Render(writer http.ResponseWriter) error {
	if header := writer.Header(); len(header.Get("Content-Type")) == 0 {
		contentType := "application/octet-stream"
		if len(b.ContentType) > 0 {
			contentType = b.ContentType
		}
		header.Set("Content-Type", contentType)
	}
	_, err := writer.Write(b.Data)
	return err
}
