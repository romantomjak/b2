package testutil

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type FakeHTTPClient struct {
	Request  *http.Request
	Response *http.Response
	Error    error
}

func (f *FakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	f.Response.Request = req
	return f.Response, f.Error
}

func HTTPResponse(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}
}
