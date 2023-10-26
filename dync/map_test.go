package dync

import (
	"encoding/json"
	"testing"

	"github.com/go-spring-projects/go-spring/conf"
	"github.com/go-spring-projects/go-spring/internal/utils/assert"
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

	p := assert.Must(conf.Map(nil))
	err := u.OnRefresh(p, param)
	assert.Equal(t, err, nil)

	_ = p.Set("map.a", "A")
	_ = p.Set("map.b", "B")
	_ = p.Set("map.c", "C")
	_ = p.Set("map.d.e.a", "E")
	_ = p.Set("map.d.f.b", "F")
	_ = p.Set("map.d.g.c", "G")

	param.Validate = ""
	err = u.OnRefresh(p, param)
	assert.Equal(t, u.Value(), map[string]string{"a": "A", "b": "B", "c": "C", "d.e.a": "E", "d.f.b": "F", "d.g.c": "G"})

	b, err := json.Marshal(&u)
	assert.Nil(t, err)
	assert.Equal(t, string(b), `{"a":"A","b":"B","c":"C","d.e.a":"E","d.f.b":"F","d.g.c":"G"}`)
}
