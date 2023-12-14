/*
 * Copyright 2019 the original author or authors.
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

package gs

import (
	"testing"

	"go-spring.dev/spring/conf"
	"go-spring.dev/spring/internal/utils/assert"
)

func TestLoadCmdArgs(t *testing.T) {
	t.Run("", func(t *testing.T) {
		err := loadCmdArgs([]string{"-D"}, nil)
		assert.Error(t, err, "cmd option -D needs arg")
	})
	t.Run("", func(t *testing.T) {
		p := conf.New()
		err := loadCmdArgs([]string{
			"-D", "language=go",
			"-D", "server",
		}, p)
		assert.Nil(t, err)
		assert.Equal(t, p.Keys(), []string{"language", "server"})
		assert.Equal(t, p.Get("language"), "go")
		assert.Equal(t, p.Get("server"), "true")
	})
}

func TestConvertEnv(t *testing.T) {

	var cases = []struct {
		key    string
		expect string
	}{
		{"a.b.c", EnvPrefix + "A_B_C"},
		{"a", EnvPrefix + "A"},
		{"a.b.c.d", EnvPrefix + "A_B_C_D"},
		{"http.port", EnvPrefix + "HTTP_PORT"},
		{"web-server.listen", EnvPrefix + "WEB-SERVER_LISTEN"},
		{"_a_b_c", EnvPrefix + "_A_B_C"},
		{"go-spring.name", EnvPrefix + "GO-SPRING_NAME"},
		{"gs_spring.level", EnvPrefix + "SPRING_LEVEL"},
		{"GS_PROJECT_NAME", EnvPrefix + "PROJECT_NAME"},
	}

	for _, c := range cases {
		assert.Equal(t, convertToEnv(c.key), c.expect)
	}
}
