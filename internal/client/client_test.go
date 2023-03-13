package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func TestNewClient(t *testing.T) {
	require := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal("/sabnzbd/api", r.URL.Path)
		require.Equal("abc123", r.URL.Query().Get("apikey"))
		require.Equal("json", r.URL.Query().Get("output"))
		require.Equal("queue", r.URL.Query().Get("mode"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := NewSabnzbdClient(ts.URL, "abc123")
	require.NoError(err)
	require.NotNil(client)
	_, err = client.Get("queue")
	require.NoError(err)
}

func TestNewClientWithInvalidURL(t *testing.T) {
	require := require.New(t)

	client, err := NewSabnzbdClient("", "abc123")
	require.Error(err)
	require.Nil(client)
}

func TestGet_GoodStatusCodes(t *testing.T) {
	require := require.New(t)
	parameters := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
	}

	for _, parameter := range parameters {
		t.Run(fmt.Sprintf("%d", parameter), func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(parameter)
			}))
			defer ts.Close()

			client, err := NewSabnzbdClient(ts.URL, "abc123")
			require.NoError(err)
			require.NotNil(client)
			resp, err := client.Get("queue")
			require.NoError(err)
			require.NotNil(resp)
		})
	}
}

func TestGet_BadStatusCodes(t *testing.T) {
	require := require.New(t)
	parameters := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	}

	for _, parameter := range parameters {
		t.Run(fmt.Sprintf("%d", parameter), func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(parameter)
			}))
			defer ts.Close()

			client, err := NewSabnzbdClient(ts.URL, "abc123")
			require.NoError(err)
			require.NotNil(client)
			_, err = client.Get("queue")
			require.Error(err)
		})
	}
}
