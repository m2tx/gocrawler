package collector

import (
	"io"
	"net/http"
	"strings"
)

type HTTPClientMock struct {
	StatusCode int
	Body       string
	DoCount    int
}

func (mock *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	mock.DoCount++

	bodyReader := io.NopCloser(strings.NewReader(mock.Body))

	return &http.Response{
		Body:       bodyReader,
		StatusCode: mock.StatusCode,
	}, nil
}
