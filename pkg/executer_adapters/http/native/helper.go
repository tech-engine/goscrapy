package nativeadapter

import (
	"net/http"

	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

func NativeHTTPRequestAdapterResponse(target executorhttp.ResponseWriter, source *http.Response, err error) error {
	if err != nil {
		return err
	}

	target.SetHeaders(source.Header).SetStatusCode(source.StatusCode).SetBody(source.Body)
	return nil
}
