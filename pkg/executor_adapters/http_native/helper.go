package httpnative

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func HTTPRequestAdapterResponse(res core.IResponseWriter, source *http.Response) {

	res.WriteHeader(source.Header)
	res.WriteStatusCode(source.StatusCode)
	res.WriteCookies(source.Cookies())
	res.WriteBody(source.Body)
}
