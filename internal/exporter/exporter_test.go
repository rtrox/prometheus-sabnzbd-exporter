package exporter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

const API_KEY = "abcdef0123456789abcdef0123456789"

func newTestServer(t *testing.T, fn func(http.ResponseWriter, *http.Request)) (*httptest.Server, error) {
	queue, err := os.ReadFile("test_fixtures/queue.json")
	require.NoError(t, err)
	serverStats, err := os.ReadFile("test_fixtures/server_stats.json")
	require.NoError(t, err)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
		require.NotEmpty(t, r.URL.Query().Get("mode"))
		switch r.URL.Query().Get("mode") {
		case "queue":
			w.WriteHeader(http.StatusOK)
			w.Write(queue)
		case "server_stats":
			w.WriteHeader(http.StatusOK)
			w.Write(serverStats)
		}
	})), nil
}

func TestCollect(t *testing.T) {
	require := require.New(t)
	ts, err := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal("/sabnzbd/api", r.URL.Path)
		require.Equal(API_KEY, r.URL.Query().Get("apikey"))
		require.Equal("json", r.URL.Query().Get("output"))
	})
	require.NoError(err)

	collector, err := NewSabnzbdExporter(ts.URL, API_KEY)
	require.NoError(err)

	require.Equal(25, testutil.CollectAndCount(collector))

	b, err := os.ReadFile("test_fixtures/expected_metrics.txt")
	require.NoError(err)
	expected := strings.Replace(string(b), "http://127.0.0.1:39965", ts.URL, -1)
	f := strings.NewReader(expected)
	err = testutil.CollectAndCompare(collector, f)
	require.NoError(err)
}
