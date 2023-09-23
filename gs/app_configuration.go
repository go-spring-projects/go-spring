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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-spring-projects/go-spring/conf"
)

type Configuration struct {
	resourceLocator  ResourceLocator
	ActiveProfiles   []string `value:"${spring.config.profiles:=}"`
	ConfigExtensions []string `value:"${spring.config.extensions:=.properties,.yaml,.yml,.toml,.tml}"`
}

func NewConfiguration(resourceLocator ResourceLocator) *Configuration {
	return &Configuration{resourceLocator: resourceLocator}
}

func (e *Configuration) Load(props *conf.Properties) error {
	p := conf.New()

	if err := loadSystemEnv(p); err != nil {
		return err
	}
	if err := loadCmdArgs(os.Args, p); err != nil {
		return err
	}
	if err := p.Bind(e); err != nil {
		return err
	}
	if err := p.Bind(e.resourceLocator); err != nil {
		return err
	}

	// 从文件加载的配置
	if err := e.loadProperties(props); nil != err {
		return err
	}

	// 从环境变量和参数获取的配置优先级更高
	for _, k := range p.Keys() {
		props.Set(k, p.Get(k))
	}
	return nil
}

func (e *Configuration) loadProperties(props *conf.Properties) error {
	var resources []Resource

	for _, ext := range e.ConfigExtensions {
		sources, err := e.loadResource("application" + ext)
		if err != nil {
			return err
		}
		resources = append(resources, sources...)
	}

	for _, profile := range e.ActiveProfiles {
		for _, ext := range e.ConfigExtensions {
			sources, err := e.loadResource("application-" + profile + ext)
			if err != nil {
				return err
			}
			resources = append(resources, sources...)
		}
	}

	defer func() {
		for _, resource := range resources {
			_ = resource.Close()
		}
	}()

	for _, resource := range resources {
		b, err := ioutil.ReadAll(resource)
		if err != nil {
			return err
		}
		p, err := conf.Bytes(b, filepath.Ext(resource.Name()))
		if err != nil {
			return err
		}
		for _, key := range p.Keys() {
			props.Set(key, p.Get(key))
		}
	}

	return nil
}

func (e *Configuration) loadResource(filename string) ([]Resource, error) {

	var locators []ResourceLocator
	locators = append(locators, e.resourceLocator)

	var resources []Resource
	for _, locator := range locators {
		sources, err := locator.Locate(filename)
		if err != nil {
			return nil, err
		}
		resources = append(resources, sources...)
	}
	return resources, nil
}
