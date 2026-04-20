package gos

import (
	"net/http"

	"github.com/tech-engine/goscrapy/internal/types"
)

type appOpts struct {
	client *http.Client
}

func defaultAppOpts() appOpts {
	return appOpts{
		client: DefaultClient(),
	}
}

// WithClient allows overriding the default HTTP client used by the app.
func WithClient(client *http.Client) types.OptFunc[appOpts] {
	return func(opts *appOpts) {
		opts.client = client
	}
}
