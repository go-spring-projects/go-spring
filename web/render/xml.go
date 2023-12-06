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
	"net/http"
)

type XmlRenderer struct {
	Prefix string
	Indent string
	Data   interface{}
}

func (x XmlRenderer) ContentType() string {
	return "application/xml; charset=utf-8"
}

func (x XmlRenderer) Render(writer http.ResponseWriter) error {
	encoder := xml.NewEncoder(writer)
	if len(x.Prefix) > 0 || len(x.Indent) > 0 {
		encoder.Indent(x.Prefix, x.Indent)
	}
	return encoder.Encode(x.Data)
}
