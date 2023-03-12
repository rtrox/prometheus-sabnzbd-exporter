package client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type testRoundTripFunc func(req *http.Request) (*http.Response, error)

func (t testRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return t(req)
}

func TestRoundTrip(t *testing.T) {
	require := require.New(t)
	transport := SabnzbdTransport{
		apiKey: "abc123",
		inner: testRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			require.NotNil(req)
			require.Equal("http://localhost:8080/sabnzbd/api?apikey=abc123&output=json", req.URL.String())
			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		}),
	}
	req, err := http.NewRequest("GET", "http://localhost:8080/sabnzbd/api", nil)
	require.NotNil(req)
	require.NoError(err)
	resp, err := transport.RoundTrip(req)
	require.NotNil(resp)
	require.NoError(err)
}
