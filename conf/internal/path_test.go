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

package internal

import (
	"errors"
	"fmt"
	"testing"

	"go-spring.dev/spring/internal/utils/assert"
)

func TestSplitPath(t *testing.T) {
	var testcases = []struct {
		Key  string
		Err  error
		Path []Path
	}{
		{
			Key: "",
		},
		{
			Key: " ",
			Err: errors.New("invalid key ' '"),
		},
		{
			Key: ".",
			Err: errors.New("invalid key '.'"),
		},
		{
			Key: "[",
			Err: errors.New("invalid key '['"),
		},
		{
			Key: "]",
			Err: errors.New("invalid key ']'"),
		},
		{
			Key: "[]",
			Err: errors.New("invalid key '[]'"),
		},
		{
			Key: "[0]",
			Path: []Path{
				{PathTypeIndex, "0"},
			},
		},
		{
			Key: "[0][",
			Err: errors.New("invalid key '[0]['"),
		},
		{
			Key: "[0]]",
			Err: errors.New("invalid key '[0]]'"),
		},
		{
			Key: "[[0]]",
			Err: errors.New("invalid key '[[0]]'"),
		},
		{
			Key: "[.]",
			Err: errors.New("invalid key '[.]'"),
		},
		{
			Key: "[a.b]",
			Err: errors.New("invalid key '[a.b]'"),
		},
		{
			Key: "a",
			Path: []Path{
				{PathTypeKey, "a"},
			},
		},
		{
			Key: "a.",
			Err: errors.New("invalid key 'a.'"),
		},
		{
			Key: "a.b",
			Path: []Path{
				{PathTypeKey, "a"},
				{PathTypeKey, "b"},
			},
		},
		{
			Key: "a[",
			Err: errors.New("invalid key 'a['"),
		},
		{
			Key: "a]",
			Err: errors.New("invalid key 'a]'"),
		},
		{
			Key: "a[0]",
			Path: []Path{
				{PathTypeKey, "a"},
				{PathTypeIndex, "0"},
			},
		},
		{
			Key: "a.[0]",
			Err: errors.New("invalid key 'a.[0]'"),
		},
		{
			Key: "a[0].b",
			Path: []Path{
				{PathTypeKey, "a"},
				{PathTypeIndex, "0"},
				{PathTypeKey, "b"},
			},
		},
		{
			Key: "a.[0].b",
			Err: errors.New("invalid key 'a.[0].b'"),
		},
		{
			Key: "a[0][0]",
			Path: []Path{
				{PathTypeKey, "a"},
				{PathTypeIndex, "0"},
				{PathTypeIndex, "0"},
			},
		},
		{
			Key: "a.[0].[0]",
			Err: errors.New("invalid key 'a.[0].[0]'"),
		},
	}
	for i, c := range testcases {
		p, err := SplitPath(c.Key)
		assert.Equal(t, err, c.Err, fmt.Sprintf("index: %d key: %q", i, c.Key))
		assert.Equal(t, p, c.Path, fmt.Sprintf("index:%d key: %q", i, c.Key))
		if err == nil {
			s := JoinPath(p)
			assert.Equal(t, s, c.Key, fmt.Sprintf("index:%d key: %q", i, c.Key))
		}
	}
}
