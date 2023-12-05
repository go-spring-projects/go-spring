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
	"crypto/rand"
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestBinaryRenderer(t *testing.T) {

	data := make([]byte, 1024)
	if _, err := rand.Reader.Read(data); nil != err {
		panic(err)
	}

	w := httptest.NewRecorder()

	err := BinaryRenderer{ContentType: "application/octet-stream", Data: data}.Render(w)
	assert.Nil(t, err)

	assert.Equal(t, w.Header().Get("Content-Type"), "application/octet-stream")
	assert.Equal(t, w.Body.Bytes(), data)
}
