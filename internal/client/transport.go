package client

import (
	"compress/gzip"
	"io"
	"net/http"
)

type SabnzbdTransport struct {
	apiKey string
	inner  http.RoundTripper
}

func (t *SabnzbdTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()
	q.Add("apikey", t.apiKey)
	q.Add("output", "json")
	req.URL.RawQuery = q.Encode()
	return t.inner.RoundTrip(req)
}

type GzipTransport struct {
	inner http.RoundTripper
}

func (t *GzipTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept-Encoding", "gzip")
	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = reader
	}

	return resp, nil
}
