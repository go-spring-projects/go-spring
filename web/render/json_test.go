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
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestJSONRenderer(t *testing.T) {
	data := map[string]any{
		"foo":  "bar",
		"html": "<b>",
	}

	w := httptest.NewRecorder()

	render := JsonRenderer{Data: data}
	err := render.Render(w)
	assert.Nil(t, err)

	assert.Equal(t, render.ContentType(), "application/json; charset=utf-8")
	assert.Equal(t, w.Body.String(), "{\"foo\":\"bar\",\"html\":\"\\u003cb\\u003e\"}\n")
}
