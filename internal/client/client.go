package client

import (
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
		var repErr error
		baseURI, repErr = url.ParseRequestURI("http://" + baseURL)
		if repErr != nil {
			return nil, err
		}
	}
	baseURI = baseURI.JoinPath(BASE_URI_PATH)
	log.Info().Interface("baseURI", baseURI).Str("baseURIString", baseURI.String()).Msg("baseURI")
	return &SabnzbdClient{
		baseURI: baseURI,
		client: &http.Client{
			Transport: &GzipTransport{
				&SabnzbdTransport{
					inner:  http.DefaultTransport,
					apiKey: apiKey,
				}},
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
	return c.client.Do(req)
}
