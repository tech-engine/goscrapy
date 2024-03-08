package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func retry500Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestRetry(t *testing.T) {

	var (
		expectedRetryCnt uint8 = 3
		actualRetryCnt   uint8
	)

	client := &http.Client{
		Transport: Retry(
			WithMaxRetries(expectedRetryCnt),
			WithCb(func(r *http.Request, retry uint8) bool {
				actualRetryCnt = retry
				return true
			}),
		)(http.DefaultTransport),
	}

	testServer := httptest.NewServer(http.HandlerFunc(retry500Handler))

	req, err := http.NewRequest("GET", testServer.URL, nil)

	assert.Nil(t, err, "error creating http request")

	resp, err := client.Do(req)

	assert.Nil(t, err, "error making request")

	resp.Body.Close()

	assert.Equal(t, expectedRetryCnt, actualRetryCnt)
	testServer.Close()
}

func TestRetryWithCb(t *testing.T) {

	var (
		retryCnt         uint8 = 3
		expectedRetryCnt uint8 = 3
		actualRetryCnt   uint8
	)

	client := &http.Client{
		Transport: Retry(
			WithMaxRetries(retryCnt),
			WithCb(func(r *http.Request, retry uint8) bool {
				actualRetryCnt = retry
				return retry <= 1
			}),
		)(http.DefaultTransport),
	}

	testServer := httptest.NewServer(http.HandlerFunc(retry500Handler))

	req, err := http.NewRequest("GET", testServer.URL, nil)

	assert.Nil(t, err, "error creating http request")

	resp, err := client.Do(req)

	assert.Nil(t, err, "error making request")

	resp.Body.Close()

	assert.Less(t, actualRetryCnt, expectedRetryCnt)
	testServer.Close()
}

func TestRetryWithHttpCodes(t *testing.T) {

	var (
		retryCnt         uint8 = 3
		expectedRetryCnt uint8 = 0
		actualRetryCnt   uint8
	)

	client := &http.Client{
		Transport: Retry(
			WithMaxRetries(retryCnt),
			WithHttpCodes([]int{467}),
			WithCb(func(r *http.Request, retry uint8) bool {
				actualRetryCnt = retry
				return true
			}),
		)(http.DefaultTransport),
	}

	testServer := httptest.NewServer(http.HandlerFunc(retry500Handler))

	req, err := http.NewRequest("GET", testServer.URL, nil)

	assert.Nil(t, err, "error creating http request")

	resp, err := client.Do(req)

	assert.Nil(t, err, "error making request")

	resp.Body.Close()

	assert.Equal(t, expectedRetryCnt, actualRetryCnt)
	testServer.Close()
}
