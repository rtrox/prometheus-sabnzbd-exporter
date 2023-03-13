package client

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

var BASE_URI_PATH = "/sabnzbd/api"

type SabnzbdClient struct {
	baseURI *url.URL
	client  *http.Client
}

func NewSabnzbdClient(baseURL, apiKey string) (*SabnzbdClient, error) {
	var baseURI *url.URL

	baseURI, err := url.Parse(baseURL)
	if err != nil || baseURI.Host == "" {
		log.Warn().
			Str("baseURL", baseURL).
			Msg("baseURL is not a valid URL, trying to parse as host:port")

		var repErr error

		baseURI, repErr = url.Parse("http://" + baseURL)
		if repErr != nil || baseURI.Host == "" {
			log.Error().
				Err(repErr).
				Str("baseURL", baseURL).
				Msg("baseURL is not a valid URL or host:port")

			return nil, fmt.Errorf("baseURL is not a valid URL or host:port (%s) %w", baseURL, err)
		}
	}

	baseURI = baseURI.JoinPath(BASE_URI_PATH)

	return &SabnzbdClient{
		baseURI: baseURI,
		client: &http.Client{
			Transport: &SabnzbdTransport{
				inner:  http.DefaultTransport,
				apiKey: apiKey,
			},
		},
	}, nil
}

func (c *SabnzbdClient) Get(mode string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURI.String(), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("mode", mode)
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	switch code := resp.StatusCode; {
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return nil, fmt.Errorf("client error: %d", code)
	case code >= http.StatusInternalServerError:
		return nil, fmt.Errorf("server error: %d", code)
	}

	return resp, nil
}
