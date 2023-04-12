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

package utils

import (
	"sort"
	"testing"

	"github.com/limpo1989/go-spring/utils/assert"
)

func TestReadDirNames(t *testing.T) {
	names, err := ReadDirNames("testdata")
	assert.Nil(t, err)
	sort.Strings(names)
	assert.Equal(t, names, []string{"pkg", "pkg.go"})
	_, err = ReadDirNames("not_exists")
	assert.Error(t, err, "open not_exists: The system cannot find the file specified")
}

func TestContract(t *testing.T) {
	file := File()
	assert.Equal(t, Contract(file, -1), file)
	assert.Equal(t, Contract(file, 0), file)
	assert.Equal(t, Contract(file, 1), file)
	assert.Equal(t, Contract(file, 3), file)
	assert.Equal(t, Contract(file, 4), "...o")
	assert.Equal(t, Contract(file, 5), "...go")
	assert.Equal(t, Contract(file, 10000), file)
}
