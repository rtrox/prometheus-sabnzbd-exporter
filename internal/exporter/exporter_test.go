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
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
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
	defer ts.Close()

	collector, err := NewSabnzbdExporter(ts.URL, API_KEY)
	require.NoError(err)

	require.GreaterOrEqual(28, testutil.CollectAndCount(collector))

	b, err := os.ReadFile("test_fixtures/expected_metrics.txt")
	require.NoError(err)
	expected := strings.Replace(string(b), "http://127.0.0.1:39965", ts.URL, -1)
	f := strings.NewReader(expected)
	require.NotPanics(func() {
		err = testutil.CollectAndCompare(collector, f,
			"sabnzbd_downloaded_bytes",
			"sabnzbd_server_downloaded_bytes",
			"sabnzbd_server_articles_total",
			"sabnzbd_server_articles_success",
			"sabnzbd_info",
			"sabnzbd_paused",
			"sabnzbd_paused_all",
			"sabnzbd_pause_duration_seconds",
			"sabnzbd_disk_used_bytes",
			"sabnzbd_disk_total_bytes",
			"sabnzbd_remaining_quota_bytes",
			"sabnzbd_quota_bytes",
			"sabnzbd_article_cache_articles",
			"sabnzbd_article_cache_bytes",
			"sabnzbd_speed_bps",
			"sabnzbd_remaining_bytes",
			"sabnzbd_total_bytes",
			"sabnzbd_queue_size",
			"sabnzbd_status",
			"sabnzbd_time_estimate_seconds",
			"sabnzbd_queue_length",
		)
	})
	require.NoError(err)
}

func TestCollect_FailureDoesntPanic(t *testing.T) {
	require := require.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	collector, err := NewSabnzbdExporter(ts.URL, API_KEY)
	require.NoError(err)
	b, err := os.ReadFile("test_fixtures/expected_failure_metrics.txt")
	require.NoError(err)
	expected := strings.Replace(string(b), "http://127.0.0.1:39965", ts.URL, -1)
	f := strings.NewReader(expected)
	require.NotPanics(func() {
		err = testutil.CollectAndCompare(collector, f)
		require.Error(err)
	}, "Collecting metrics should not panic on failure")
	require.Error(err)
}
