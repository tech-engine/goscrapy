package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"slices"

	"github.com/stretchr/testify/assert"
)

// Set our custom transport middleware
var client = &http.Client{
	Transport: MultiCookieJar(http.DefaultTransport),
}

func filteredHeaders(h http.Header) http.Header {
	var newHeader = make(http.Header)
	skipHeaders := []string{"User-Agent", "Accept-Encoding", "Cookie", "Content-Length", "Date"}

	for name, value := range h {
		// we skip the default headers
		if slices.Contains(skipHeaders, name) {
			continue
		}
		newHeader.Add(name, value[0])
	}

	return newHeader
}

func makeTestRequestWithClient(client *http.Client) func(string, string, http.Header) (*http.Response, error) {
	return func(method, url string, header http.Header) (*http.Response, error) {
		// Create a first GET request without any cookie
		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			return nil, fmt.Errorf("makeRequestWithClient: error creating http request %w", err)
		}

		if header != nil {
			req.Header = header
		}

		return client.Do(req)
	}
}

// handlerGetCookieJar provides us our dummy server handlers
func handlerGetCookieJar(t *testing.T) *http.ServeMux {
	mux := http.NewServeMux()
	skipHeaders := []string{"User-Agent", "Accept-Encoding", "Cookie"}
	// /get-cookie receives headers from client and set those headers as response cookies
	mux.HandleFunc("/set-cookie", func(w http.ResponseWriter, r *http.Request) {
		for name, value := range r.Header {
			// Set the cookie in the response
			if len(value) <= 0 {
				continue
			}

			// we skip the default headers
			if slices.Contains(skipHeaders, name) {
				continue
			}

			http.SetCookie(w, &http.Cookie{
				Name:   name,
				Value:  value[0],
				Domain: r.URL.Host,
				Path:   "/",
			})
		}

		w.WriteHeader(http.StatusOK)

	})

	// /verify receives cookies(auto injected by middleware) from client and respond back with the cookie value in response headers
	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {

		// Inspect the request's cookies
		receivedCookies := r.Cookies()

		for _, c := range receivedCookies {
			w.Header().Set(c.Name, c.Value)
		}

		w.WriteHeader(http.StatusOK)

	})
	return mux
}

// There are 2 stages.
//
// Stage 1: We send a request with a few headers to our test server, and get the exact same headers back
// as response cookies.
//
// Stage 2: To verify if our cookie middleware worked as expected, we will send another request to /verify.
// If we get the exact same headers we sent in Stage 1, as response headers, our middleware worked as expected.
func RunWithCookieJar(t *testing.T, key string) {
	// Create a test server with the custom RoundTripper
	key = strings.ToLower(key)
	testServer := httptest.NewServer(handlerGetCookieJar(t))
	defer testServer.Close()

	requester := makeTestRequestWithClient(client)

	// Stage 1
	headerOne := http.Header{
		"X-Goscrapy-Server-Req-" + key: []string{"single_host_req_" + key},
	}

	if key != "" {
		headerOne.Add("X-Goscrapy-Cookie-Jar-Key", key)
	}

	respOne, err := requester("GET", testServer.URL+"/set-cookie", headerOne)

	assert.Nil(t, err, "error making http request 1")

	defer respOne.Body.Close()

	// we verify if we have received the same cookies that we have set in "X-Goscrapy-Server-Req-1" header
	respOneCookies := respOne.Cookies()

	assert.Lenf(t, respOneCookies, 1, "expected only %d cookie but got %d", 1, len(respOneCookies))

	found := false

	for _, cookie := range respOneCookies {
		if strings.ToLower(cookie.Name) == "x-goscrapy-server-req-"+key && strings.ToLower(cookie.Value) == "single_host_req_"+key {
			found = true
			break
		}
	}

	assert.Truef(t, found, "expected response cookies [X-Goscrapy-Server-Req-%s=single_host_req_%s] not found", key, key)

	// second stage 2:
	headerTwo := http.Header{
		"X-Goscrapy-Cookie-Jar-Key": []string{key},
	}
	respTwo, err := requester("GET", testServer.URL+"/verify", headerTwo)

	assert.Nil(t, err, "error making http request 2")

	defer respTwo.Body.Close()

	respTwoHeader := filteredHeaders(respTwo.Header)

	assert.Lenf(t, respTwoHeader, 1, "expected only %d header but got %d", 1, len(respTwoHeader))

	assert.Equal(t, "single_host_req_"+key, respTwoHeader.Get("X-Goscrapy-Server-Req-"+key))
}

func TestMultiCookierJar(t *testing.T) {

	testCases := []struct {
		Name,
		SessionKey string
	}{
		{
			Name:       "DEFAULT_COOKIEJAR",
			SessionKey: "",
		},
		{
			Name:       "SINGLE_COOKIEJAR",
			SessionKey: "jar1",
		},
		{
			Name:       "SINGLE_COOKIEJAR",
			SessionKey: "jar2",
		},
		{
			Name:       "SINGLE_COOKIEJAR",
			SessionKey: "jar3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			RunWithCookieJar(t, tc.SessionKey)
		})
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			RunWithCookieJar(t, tc.SessionKey)
		})
	}
}
