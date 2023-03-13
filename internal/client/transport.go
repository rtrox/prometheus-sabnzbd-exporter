package client

import (
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
