package httpnative

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

func HTTPRequestAdapterResponse(res engine.IResponseWriter, source *http.Response, err error) error {
	if err != nil {
		return err
	}

	res.WriteHeader(source.Header)
	res.WriteStatusCode(source.StatusCode)
	res.WriteCookies(source.Cookies())
	res.WriteBody(source.Body)

	return nil
}
