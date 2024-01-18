package nativeadapter

import (
	"net/http"

	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

func HTTPRequestAdapterResponse(target executorhttp.ResponseSetter, source *http.Response, err error) error {
	if err != nil {
		return err
	}

	target.SetHeaders(source.Header).
		SetStatusCode(source.StatusCode).
		SetBody(source.Body).
		SetCookies(source.Cookies())
	return nil
}
