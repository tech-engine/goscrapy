package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"slices"

	"github.com/stretchr/testify/assert"
)

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

/*
Handler func to test default cookiejar.
We send 2 consecutive http requests. In first request, the server responds back with set-cookie header,
and in the second request our client sends the cookies from default cookie jar to the server.
*/
func handlerGetCookieJar(t *testing.T) *http.ServeMux {
	mux := http.NewServeMux()
	skipHeaders := []string{"User-Agent", "Accept-Encoding", "Cookie"}
	// /get-cookie responds back to client with set-cookie header
	mux.HandleFunc("/get-cookie", func(w http.ResponseWriter, r *http.Request) {

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

	// /send-cookie receives cookie from client and respond back with the cookie value in response
	mux.HandleFunc("/send-cookie", func(w http.ResponseWriter, r *http.Request) {

		// Inspect the request's cookies
		receivedCookies := r.Cookies()

		for _, c := range receivedCookies {
			w.Header().Set(c.Name, c.Value)
		}
		// add extra cookies
		for name, value := range r.Header {
			// Set the cookie in the response
			if len(value) <= 0 {
				continue
			}

			// we skip the default headers
			if slices.Contains(skipHeaders, name) {
				continue
			}

			cookie := &http.Cookie{
				Name:   name,
				Value:  value[0],
				Domain: r.URL.Host,
				Path:   "/",
			}
			http.SetCookie(w, cookie)

			w.Header().Set(name, value[0])
		}
		w.WriteHeader(http.StatusOK)

	})
	return mux
}

// Both the 2 http request to the same server HOST:PORT
func singleHostCookieJar(client *http.Client, session string) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a test server with the custom RoundTripper
		testServer := httptest.NewServer(handlerGetCookieJar(t))
		defer testServer.Close()

		requester := makeTestRequestWithClient(client)

		headerOne := http.Header{
			"X-Goscrapy-Server-Req-1": []string{"single_host_req_1"},
		}

		if session != "" {
			headerOne.Add("cookie-jar", session)
		}
		// Create the first GET request with default cookiejar, this is set the cookie in our client
		respOne, err := requester("GET", testServer.URL+"/get-cookie", headerOne)

		assert.Nil(t, err, "error making http request 1")

		defer respOne.Body.Close()

		// we verify if we have received the same cookies that we have set in "X-Goscrapy-Server-Req-1" header
		cookiesFromRespOne := respOne.Cookies()

		assert.Len(t, cookiesFromRespOne, 1, fmt.Sprintf("expected only 1 cookie but got %d", len(cookiesFromRespOne)))

		found := false
		for _, cookie := range cookiesFromRespOne {
			if cookie.Name == "X-Goscrapy-Server-Req-1" && cookie.Value == "single_host_req_1" {
				found = true
				break
			}
		}

		assert.True(t, found, "expected cookie X-Goscrapy-Server-Req-1=single_host_req_1 in respOne.Cookies(), but not found")

		headerTwo := http.Header{
			"X-Goscrapy-Server-Req-2": []string{"single_host_req_2"},
		}

		if session != "" {
			headerTwo.Add("cookie-jar", session)
		}

		// Create the second GET request with default cookiejar
		respTwo, err := requester("GET", testServer.URL+"/send-cookie", headerTwo)

		assert.Nil(t, err, "error in making http request 2")

		defer respTwo.Body.Close()

		// We read cookies from second request
		cookiesFromRespTwo := respTwo.Cookies()

		// as we set only 1 header in our second request we must receive only 1 cookie
		assert.Len(t, cookiesFromRespTwo, 1, fmt.Sprintf("expected only 1 cookie but got %d", len(cookiesFromRespTwo)))

		// check if we received the correct cookie
		found = false
		for _, cookie := range cookiesFromRespTwo {
			if cookie.Name == "X-Goscrapy-Server-Req-2" && cookie.Value == "single_host_req_2" {
				found = true
				break
			}
		}

		assert.True(t, found, "expected cookie X-Goscrapy-Server-Req-2=single_host_req_2 in respTwo.Cookies(), but not found")

		found = false
		respTwoHeaders := respTwo.Header

		// we check if we have received back the 2 cookies as response headers

		if (respTwoHeaders.Get("X-Goscrapy-Server-Req-1") == "single_host_req_1") && (respTwoHeaders.Get("X-Goscrapy-Server-Req-2") == "single_host_req_2") {
			found = true
		}

		assert.True(t, found, "expected 2 cookies, X-Goscrapy-Server-Req-1=single_host_req_1 & X-Goscrapy-Server-Req-2=single_host_req_2 to be passed to /send-cookie, but not passed")

	}
}

func TestMultiCookierJar(t *testing.T) {

	// Set our custom transport middleware
	client := &http.Client{
		Transport: MultiCookieJar(http.DefaultTransport),
	}

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
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			singleHostCookieJar(client, tc.SessionKey)(t)
		})
	}
}
