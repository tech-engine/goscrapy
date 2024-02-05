package engine

import (
	"net/http"
)

type Middleware func(next http.RoundTripper) http.RoundTripper
