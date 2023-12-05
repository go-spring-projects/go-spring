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
