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
	"errors"
	"os"
	"strings"

	"go-spring.dev/spring/conf"
)

// EnvPrefix environment variable prefix。
const EnvPrefix = "GS_"

// loadCmdArgs 加载以 -D key=value 或者 -D key[=true] 形式传入的命令行参数。
func loadCmdArgs(args []string, p *conf.Properties) error {
	for i := 0; i < len(args); i++ {
		s := args[i]
		if s == "-D" {
			if i >= len(args)-1 {
				return errors.New("cmd option -D needs arg")
			}
			next := args[i+1]
			ss := strings.SplitN(next, "=", 2)
			if len(ss) == 1 {
				ss = append(ss, "true")
			}
			if err := p.Set(ss[0], ss[1]); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadSystemEnv(p *conf.Properties) error {
	for _, env := range os.Environ() {
		ss := strings.SplitN(env, "=", 2)
		k, v := ss[0], ""
		if len(ss) > 1 {
			v = ss[1]
		}
		if strings.HasPrefix(k, EnvPrefix) {
			propKey := strings.TrimPrefix(k, EnvPrefix)
			propKey = strings.ReplaceAll(propKey, "_", ".")
			propKey = strings.ToLower(propKey)
			p.Set(propKey, v)
		}
	}
	return nil
}

func convertToEnv(key string) string {
	if strings.Contains(key, ".") {
		// replace '.' to '_'
		key = strings.ReplaceAll(key, ".", "_")
	}

	// upper case
	key = strings.ToUpper(key)

	if !strings.HasPrefix(key, EnvPrefix) {
		// append prefix
		key = EnvPrefix + key
	}
	return key
}
