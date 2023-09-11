package dync

import (
	"encoding/json"
	"testing"

	"github.com/limpo1989/go-spring/conf"
	"github.com/limpo1989/go-spring/utils/assert"
)

func TestMap(t *testing.T) {

	var u Map[string, string]
	assert.Equal(t, u.Value(), (map[string]string)(nil))

	param := conf.BindParam{
		Key:  "map",
		Path: "map",
		Tag: conf.ParsedTag{
			Key: "map",
		},
	}

	p := conf.Map(nil)
	err := u.OnRefresh(p, param)
	assert.Equal(t, err, nil)

	_ = p.Set("map.a", "A")
	_ = p.Set("map.b", "B")
	_ = p.Set("map.c", "C")

	param.Validate = ""
	err = u.OnRefresh(p, param)
	assert.Equal(t, u.Value(), map[string]string{"a": "A", "b": "B", "c": "C"})

	b, err := json.Marshal(&u)
	assert.Nil(t, err)
	assert.Equal(t, string(b), `{"a":"A","b":"B","c":"C"}`)
}
