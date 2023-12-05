package render

import (
	"crypto/rand"
	"net/http/httptest"
	"testing"

	"github.com/go-spring-projects/go-spring/internal/utils/assert"
)

func TestBinaryRenderer(t *testing.T) {

	data := make([]byte, 1024)
	if _, err := rand.Reader.Read(data); nil != err {
		panic(err)
	}

	w := httptest.NewRecorder()

	err := BinaryRenderer{ContentType: "application/octet-stream", Data: data}.Render(w)
	assert.Nil(t, err)

	assert.Equal(t, w.Header().Get("Content-Type"), "application/octet-stream")
	assert.Equal(t, w.Body.Bytes(), data)
}
