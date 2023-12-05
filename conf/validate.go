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

package conf

import (
	"fmt"
	"reflect"

	"github.com/expr-lang/expr"
)

var validators = map[string]Validator{
	"expr": &exprValidator{},
}

// Validator is interface for validating a field.
type Validator interface {
	Field(tag string, i interface{}) error
}

// Register registers a Validator with tag name.
func Register(name string, v Validator) {
	validators[name] = v
}

// Validate validates a single variable.
func Validate(tag reflect.StructTag, i interface{}) error {
	if len(tag) > 0 {
		for name, v := range validators {
			if s, ok := tag.Lookup(name); ok && s != "-" {
				if err := v.Field(s, i); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ValidateStruct validates a single struct.
func ValidateStruct(s interface{}) error {
	sValue := reflect.ValueOf(s)
	if reflect.Ptr == sValue.Type().Kind() {
		// ignore validate nil value
		if sValue.IsNil() {
			return nil
		}
		sValue = sValue.Elem()
	}
	return validStruct(sValue)
}

func validStruct(v reflect.Value) error {
	vType := v.Type()
	if reflect.Struct != vType.Kind() {
		return fmt.Errorf("%s: is not a struct", vType.String())
	}

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := vType.Field(i)
		if !fieldType.IsExported() {
			continue
		}

		if err := Validate(fieldType.Tag, fieldVal.Interface()); nil != err {
			return err
		}

		if reflect.Struct == fieldType.Type.Kind() {
			if err := validStruct(fieldVal); nil != err {
				return err
			}
		} else if reflect.Ptr == fieldType.Type.Kind() && reflect.Struct == fieldType.Type.Elem().Kind() && !fieldVal.IsNil() {
			if err := validStruct(fieldVal.Elem()); nil != err {
				return err
			}
		}
	}

	return nil
}

type exprValidator struct{}

// Field validates a single variable.
func (d exprValidator) Field(tag string, i interface{}) error {
	r, err := expr.Eval(tag, map[string]interface{}{"$": i})
	if err != nil {
		return fmt.Errorf("eval %q returns: %w", tag, err)
	}
	b, ok := r.(bool)
	if !ok {
		return fmt.Errorf("eval %q doesn't return bool", tag)
	}

	if !b {
		return fmt.Errorf("validate failed on %q for value %v", tag, i)
	}
	return nil
}
