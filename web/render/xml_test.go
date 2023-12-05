package render

import (
	"encoding/xml"
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

type xmlmap map[string]any

// Allows type H to be used with xml.Marshal
func (h xmlmap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func TestXmlRenderer(t *testing.T) {
	w := httptest.NewRecorder()
	data := xmlmap{
		"foo": "bar",
	}

	err := (XmlRenderer{Data: data}).Render(w)

	assert.Nil(t, err)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/xml; charset=utf-8")
	assert.Equal(t, w.Body.String(), "<map><foo>bar</foo></map>")
}
