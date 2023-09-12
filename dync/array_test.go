package dync

import (
	"encoding/json"
	"testing"

	"github.com/limpo1989/go-spring/conf"
	"github.com/limpo1989/go-spring/internal/utils/assert"
)

func TestArray(t *testing.T) {

	var u Array[int32]
	assert.Equal(t, u.Value(), []int32{})

	param := conf.BindParam{
		Key:  "int",
		Path: "int32",
		Tag: conf.ParsedTag{
			Key: "int",
		},
	}

	p := conf.Map(nil)
	err := u.OnRefresh(p, param)
	assert.Error(t, err, "bind \\[\\]int32 error: property \"int\": not exist")

	_ = p.Set("int[0]", int32(3))
	_ = p.Set("int[1]", int32(4))
	_ = p.Set("int[2]", int32(5))

	param.Validate = ""
	err = u.OnRefresh(p, param)
	assert.Equal(t, u.Value(), []int32{3, 4, 5})

	b, err := json.Marshal(&u)
	assert.Nil(t, err)
	assert.Equal(t, string(b), "[3,4,5]")
}
