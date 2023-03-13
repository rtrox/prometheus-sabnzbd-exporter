package models

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusToString(t *testing.T) {
	require := require.New(t)
	require.Equal("Downloading", DOWNLOADING.String())
	require.Equal("Paused", PAUSED.String())
	require.Equal("Idle", IDLE.String())
	require.Equal("Unknown", Status(999).String())
}

func TestStatusFromString(t *testing.T) {
	require := require.New(t)
	require.Equal(DOWNLOADING, StatusFromString("Downloading"))
	require.Equal(PAUSED, StatusFromString("Paused"))
	require.Equal(IDLE, StatusFromString("Idle"))
	require.Equal(UNKNOWN, StatusFromString("Unknown"))
	require.Equal(UNKNOWN, StatusFromString("Unknown"))
}

func TestStatusToFloat(t *testing.T) {
	require := require.New(t)
	require.Equal(3.0, DOWNLOADING.Float64())
	require.Equal(2.0, PAUSED.Float64())
	require.Equal(1.0, IDLE.Float64())
	require.Equal(0.0, UNKNOWN.Float64())
}

func TestNewServerStatsFromResponse(t *testing.T) {
	require := require.New(t)
	responseRaw := `{
		"total": 123456789,
		"servers": {
			"server1": {
				"total": 234567890,
				"articles_tried": {
					"2020-01-01": 1,
					"2020-01-02": 2
				},
				"articles_success": {
					"2020-01-01": 3,
					"2020-01-02": 4
				}
			},
			"server2": {
				"total": 345678901,
				"articles_tried": {
					"2020-01-01": 5,
					"2020-01-02": 6
				},
				"articles_success": {
					"2020-01-01": 7,
					"2020-01-02": 8
				}
			}
		}
	} `
	statsResponse := ServerStatsResponse{}
	err := json.NewDecoder(strings.NewReader(responseRaw)).Decode(&statsResponse)
	require.NoError(err)

	stats := NewServerStatsFromResponse(statsResponse)
	require.Equal(123456789, stats.Total)
	require.Equal(2, len(stats.Servers))
	require.Equal(234567890, stats.Servers["server1"].Total)
	require.Equal(2, stats.Servers["server1"].ArticlesTried)
	require.Equal(4, stats.Servers["server1"].ArticlesSuccess)
	require.Equal("2020-01-02", stats.Servers["server1"].DayParsed)
	require.Equal(345678901, stats.Servers["server2"].Total)
	require.Equal(6, stats.Servers["server2"].ArticlesTried)
	require.Equal(8, stats.Servers["server2"].ArticlesSuccess)
	require.Equal("2020-01-02", stats.Servers["server2"].DayParsed)
}
