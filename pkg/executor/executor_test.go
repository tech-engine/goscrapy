package executor

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tech-engine/goscrapy/internal/fsmap"
	"github.com/tech-engine/goscrapy/pkg/core"
)

// MockExecutorAdapter implements IExecutorAdapter
type MockExecutorAdapter struct {
	mock.Mock
}

func (m *MockExecutorAdapter) Do(res core.IResponseWriter, req *http.Request) error {
	args := m.Called(res, req)
	return args.Error(0)
}

func (m *MockExecutorAdapter) WithClient(client *http.Client) {
	m.Called(client)
}

func (m *MockExecutorAdapter) WithLogger(l core.ILogger) IExecutorAdapter {
	args := m.Called(l)
	return args.Get(0).(IExecutorAdapter)
}

// MockRequestReader implements IRequestReader
type MockRequestReader struct {
	mock.Mock
}

func (m *MockRequestReader) ReadUrl() *url.URL {
	args := m.Called()
	return args.Get(0).(*url.URL)
}

func (m *MockRequestReader) ReadContext() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockRequestReader) ReadMethod() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRequestReader) ReadBody() io.ReadCloser {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(io.ReadCloser)
}

func (m *MockRequestReader) ReadHeader() http.Header {
	args := m.Called()
	return args.Get(0).(http.Header)
}

func (m *MockRequestReader) ReadCookieJar() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRequestReader) ReadMeta() *fsmap.FixedSizeMap[string, any] {
	args := m.Called()
	return args.Get(0).(*fsmap.FixedSizeMap[string, any])
}

// MockResponseWriter implements IResponseWriter
type MockResponseWriter struct {
	mock.Mock
}

func (m *MockResponseWriter) WriteHeader(h http.Header) {
	m.Called(h)
}

func (m *MockResponseWriter) WriteBody(r io.ReadCloser) {
	m.Called(r)
}

func (m *MockResponseWriter) WriteStatusCode(code int) {
	m.Called(code)
}

func (m *MockResponseWriter) WriteCookies(cookies []*http.Cookie) {
	m.Called(cookies)
}

func (m *MockResponseWriter) WriteRequest(req *http.Request) {
	m.Called(req)
}

func (m *MockResponseWriter) WriteMeta(meta *fsmap.FixedSizeMap[string, any]) {
	m.Called(meta)
}

func TestExecutor_Execute(t *testing.T) {
	mockAdapter := new(MockExecutorAdapter)
	exec := New(mockAdapter)

	u, _ := url.Parse("http://localhost")
	ctx := context.Background()
	header := http.Header{"User-Agent": []string{"test"}}

	mockReq := new(MockRequestReader)
	mockReq.On("ReadUrl").Return(u)
	mockReq.On("ReadContext").Return(ctx)
	mockReq.On("ReadMethod").Return("POST")
	mockReq.On("ReadBody").Return(nil)
	mockReq.On("ReadHeader").Return(header)
	mockReq.On("ReadCookieJar").Return("session1")

	mockRes := new(MockResponseWriter)

	mockAdapter.On("Do", mockRes, mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == "POST" && 
			req.URL.String() == "http://localhost" &&
			req.Header.Get("User-Agent") == "test" &&
			core.ExtractCtxValue(req.Context(), "GOSCookieJarKey") == "session1"
	})).Return(nil)

	err := exec.Execute(mockReq, mockRes)
	assert.NoError(t, err)

	mockAdapter.AssertExpectations(t)
	mockReq.AssertExpectations(t)
}
