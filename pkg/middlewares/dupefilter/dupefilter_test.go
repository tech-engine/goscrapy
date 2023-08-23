package dupefilter

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	Name,
	Method string
	Header   http.Header
	Body     io.Reader
	mayBlock bool
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestDupeFilter(t *testing.T) {

	// Set our custom transport middleware
	client := &http.Client{
		Transport: DupeFilterMiddleware(http.DefaultTransport),
	}

	testServer := httptest.NewServer(http.HandlerFunc(handler))

	testCases := []testCase{
		{
			Name:   "test1",
			Method: "GET",
			Header: http.Header{
				"X-test": []string{"test_val_1"},
			},
			mayBlock: true,
		},
		{
			Name:   "test2",
			Method: "GET",
			Header: http.Header{
				"X-test": []string{"test_val_2"},
			},
			Body: nil,
		},
		{
			Name:   "test3",
			Method: "GET",
			Header: http.Header{
				"X-test": []string{"test_val_1"},
			},
			Body:     nil,
			mayBlock: true,
		},
		{
			Name:   "test4",
			Method: "POST",
			Header: http.Header{
				"X-test-another": []string{"test_val_1"},
			},
			Body:     strings.NewReader("hello"),
			mayBlock: true,
		},
		{
			Name:   "test5",
			Method: "PATCH",
			Header: http.Header{
				"X-test-another": []string{"test_val_1"},
			},
			Body: strings.NewReader("hello"),
		},
		{
			Name:   "test6",
			Method: "POST",
			Header: http.Header{
				"X-test-another": []string{"test_val_1"},
			},
			Body:     strings.NewReader("hello"),
			mayBlock: true,
		},
		{
			Name:   "test7",
			Method: "POST",
			Header: http.Header{
				"X-test-another": []string{"test_val_1"},
			},
			Body: strings.NewReader("hello1"),
		},
	}

	expectedPassCases := 5
	actualPassCases := 0

	var m sync.Mutex
	for _, tc := range testCases {
		func(tc testCase) {

			t.Run(tc.Method, func(t *testing.T) {
				t.Parallel()
				req, err := http.NewRequest(tc.Method, testServer.URL, tc.Body)

				assert.Nil(t, err, "error creating http request", tc.Name)

				req.Header = tc.Header

				resp, err := client.Do(req)
				if tc.mayBlock && err != nil {
					assert.ErrorIs(t, err, ERR_DUPEFILTER_BLOCKED, fmt.Sprintf("http request %s not blocked", tc.Name))
				}

				if resp != nil {
					assert.Equal(t, 200, resp.StatusCode, "statuscode 200 expected")
					m.Lock()
					defer m.Unlock()
					actualPassCases++
				}

			})
		}(tc)

	}

	t.Cleanup(func() {
		assert.Equal(t, expectedPassCases, actualPassCases, fmt.Sprintf("expected pass cases %d not equal to actual pass cases %d", expectedPassCases, actualPassCases))
		testServer.Close()
	})

}
