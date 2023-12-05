package render

import (
	"fmt"
	"net/http"
)

type RedirectRenderer struct {
	Code     int
	Request  *http.Request
	Location string
}

func (r RedirectRenderer) Render(writer http.ResponseWriter) error {
	if (r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
		panic(fmt.Sprintf("Cannot redirect with status code %d", r.Code))
	}
	http.Redirect(writer, r.Request, r.Location, r.Code)
	return nil
}
