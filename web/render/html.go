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
	"html/template"
	"net/http"
)

type HTMLRenderer struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

func (h HTMLRenderer) ContentType() string {
	return "text/html; charset=utf-8"
}

func (h HTMLRenderer) Render(writer http.ResponseWriter) error {
	if len(h.Name) > 0 {
		return h.Template.ExecuteTemplate(writer, h.Name, h.Data)
	}
	return h.Template.Execute(writer, h.Data)
}
