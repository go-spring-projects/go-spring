/*
 * Copyright 2023 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package render

import (
	"encoding/xml"
	"net/http/httptest"
	"testing"

	"go-spring.dev/spring/internal/utils/assert"
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

	render := (XmlRenderer{Data: data})
	err := render.Render(w)

	assert.Nil(t, err)
	assert.Equal(t, render.ContentType(), "application/xml; charset=utf-8")
	assert.Equal(t, w.Body.String(), "<map><foo>bar</foo></map>")
}
