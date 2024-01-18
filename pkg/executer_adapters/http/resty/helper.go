package restyadapter

import (
	"github.com/go-resty/resty/v2"
	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

func HTTPRequestAdapterResponse(target executorhttp.ResponseSetter, source *resty.Response, err error) error {
	if err != nil {
		return err
	}

	target.SetHeaders(source.Header()).
		SetStatusCode(source.StatusCode()).
		SetBody(source.RawResponse.Body).
		SetCookies(source.Cookies())
	return nil
}
