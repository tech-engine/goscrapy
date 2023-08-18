package restyadapter

import (
	"github.com/go-resty/resty/v2"
	executorhttp "github.com/tech-engine/goscrapy/internal/executer/http"
)

func RestyHTTPRequestAdapterResponse(target executorhttp.ResponseWriter, source *resty.Response, err error) error {
	if err != nil {
		return err
	}

	// target.SetHeaders(source.Header()).SetStatusCode(source.StatusCode()).SetBody(source.Body())
	return nil
}
