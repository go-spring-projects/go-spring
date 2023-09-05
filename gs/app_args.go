/*
 * Copyright 2012-2019 the original author or authors.
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
	"regexp"
	"strings"

	"github.com/limpo1989/go-spring/conf"
)

// EnvPrefix 属性覆盖的环境变量需要携带该前缀。
const EnvPrefix = "GS_"

// IncludeEnvPatterns 只加载符合条件的环境变量。
const IncludeEnvPatterns = "INCLUDE_ENV_PATTERNS"

// ExcludeEnvPatterns 排除符合条件的环境变量。
const ExcludeEnvPatterns = "EXCLUDE_ENV_PATTERNS"

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

	toRex := func(patterns []string) ([]*regexp.Regexp, error) {
		var rex []*regexp.Regexp
		for _, v := range patterns {
			exp, err := regexp.Compile(v)
			if err != nil {
				return nil, err
			}
			rex = append(rex, exp)
		}
		return rex, nil
	}

	includes := []string{".*"}
	if s, ok := os.LookupEnv(IncludeEnvPatterns); ok {
		includes = strings.Split(s, ",")
	}
	includeRex, err := toRex(includes)
	if err != nil {
		return err
	}

	var excludes []string
	if s, ok := os.LookupEnv(ExcludeEnvPatterns); ok {
		excludes = strings.Split(s, ",")
	}
	excludeRex, err := toRex(excludes)
	if err != nil {
		return err
	}

	matches := func(rex []*regexp.Regexp, s string) bool {
		for _, r := range rex {
			if r.MatchString(s) {
				return true
			}
		}
		return false
	}

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
			continue
		}
		if matches(includeRex, k) && !matches(excludeRex, k) {
			p.Set(k, v)
		}
	}
	return nil
}
