package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert := assert.New(t)
	statsResponse := ServerStatsResponse{
		Total: 123456789,
		Servers: map[string]ServerStatResponse{
			"server1": {
				Total: 234567890,
				ArticlesTried: map[string]int{
					"2020-01-01": 1,
					"2020-01-02": 2,
				},
				ArticlesSuccess: map[string]int{
					"2020-01-01": 3,
					"2020-01-02": 4,
				},
			},
			"server2": {
				Total: 345678901,
				ArticlesTried: map[string]int{
					"2020-01-02": 6,
					"2020-01-01": 5,
				},
				ArticlesSuccess: map[string]int{
					"2020-01-02": 8,
					"2020-01-01": 7,
				},
			},
		},
	}
	stats := NewServerStatsFromResponse(statsResponse)
	assert.Equal(123456789, stats.Total)
	assert.Equal(2, len(stats.Servers))
	assert.Equal(234567890, stats.Servers["server1"].Total)
	assert.Equal(2, stats.Servers["server1"].ArticlesTried)
	assert.Equal(4, stats.Servers["server1"].ArticlesSuccess)
	assert.Equal("2020-01-02", stats.Servers["server1"].DayParsed)
	assert.Equal(345678901, stats.Servers["server2"].Total)
	assert.Equal(6, stats.Servers["server2"].ArticlesTried)
	assert.Equal(8, stats.Servers["server2"].ArticlesSuccess)
	assert.Equal("2020-01-02", stats.Servers["server2"].DayParsed)
}

func TestNewQueueStatsFromResponse(t *testing.T) {
	assert := assert.New(t)
	statsResponse := QueueResponse{
		QueueResponseQueue{
			Version:         "3.7.2",
			Paused:          false,
			PauseInt:        "0",
			PausedAll:       false,
			Diskspace1:      "34773.60",
			Diskspace2:      "34719.60",
			DiskspaceTotal1: "42888.0",
			DiskspaceTotal2: "42889.0",
			Speedlimit:      "100",
			SpeedlimitAbs:   "1048576000",
			HaveWarnings:    "0",
			Quota:           "1005.0 G",
			HaveQuota:       true,
			LeftQuota:       "1005.0 G",
			CacheArt:        "0",
			CacheSize:       "0 B",
			KBPerSec:        "0.35",
			MBLeft:          "3061.97",
			MB:              "3062.97",
			NoofSlotsTotal:  2,
			Status:          "Downloading",
			TimeLeft:        "103:23:59:03",
		},
	}
	stats, err := NewQueueStatsFromResponse(statsResponse)
	require.NoError(t, err)
	assert.Equal("3.7.2", stats.Version)
	assert.Equal(false, stats.Paused)
	assert.Equal(time.Duration(0), stats.PauseDuration)
	assert.Equal(false, stats.PausedAll)
	assert.Equal(36462762393.6, stats.DownloadDirDiskspaceUsed)
	assert.Equal(36406139289.6, stats.CompletedDirDiskspaceUsed)
	assert.Equal(44971327488.0, stats.DownloadDirDiskspaceTotal)
	assert.Equal(44972376064.0, stats.CompletedDirDiskspaceTotal)
	assert.Equal(100.0, stats.SpeedLimit)
	assert.Equal(1048576000.0, stats.SpeedLimitAbs)
	assert.Equal(0.0, stats.HaveWarnings)
	assert.Equal(1079110533120.0, stats.Quota)
	assert.Equal(true, stats.HaveQuota)
	assert.Equal(1079110533120.0, stats.RemainingQuota)
	assert.Equal(0.0, stats.CacheArt)
	assert.Equal(0.0, stats.CacheSize)
	assert.Equal(358.4, stats.Speed)
	assert.Equal(3210708254.72, stats.RemainingSize)
	assert.Equal(3211756830.72, stats.Size)
	assert.Equal(2.0, stats.ItemsInQueue)
	assert.Equal(DOWNLOADING, stats.Status)

	expected, _ := time.ParseDuration("2495h59m3s")
	assert.Equal(expected, stats.TimeEstimate)
}

func TestNewQueueStatsFromResponse_ParsingSize(t *testing.T) {
	parameters := []struct {
		input    string
		expected float64
	}{
		{"0 B", 0.0},
		{"1 B", 1.0},
		{"1.0 B", 1.0},
		{"10 K", 10240.0},
		{"10.0 KB", 10240.0},
		{"10 M", 10485760.0},
		{"10.0 MB", 10485760.0},
		{"10 G", 10737418240.0},
		{"10.0 GB", 10737418240.0},
		{"10 T", 10995116277760.0},
		{"10.0 TB", 10995116277760.0},
		{"10 P", 11258999068426240.0},
		{"10.0 PB", 11258999068426240.0},
	}

	require := require.New(t)

	for _, parameter := range parameters {
		statsResponse := QueueResponse{
			QueueResponseQueue{
				LeftQuota: parameter.input,
			},
		}
		stats, err := NewQueueStatsFromResponse(statsResponse)
		require.NoError(err)
		require.Equal(parameter.expected, stats.RemainingQuota)
	}
}

func TestNewQueueStatsFromReponse_ParsingDuration(t *testing.T) {
	parameters := []struct {
		input    string
		expected time.Duration
	}{
		{"", time.Duration(0)},
		{"10", time.Duration(10) * time.Second},
		{"10:01", time.Duration(10)*time.Minute + time.Duration(1)*time.Second},
		{"13:12:11", time.Duration(13)*time.Hour + time.Duration(12)*time.Minute + time.Duration(11)*time.Second},
		{"14:13:12:11", time.Duration(349)*time.Hour + time.Duration(12)*time.Minute + time.Duration(11)*time.Second},
	}

	require := require.New(t)

	for _, parameter := range parameters {
		statsResponse := QueueResponse{
			QueueResponseQueue{
				TimeLeft: parameter.input,
			},
		}
		stats, err := NewQueueStatsFromResponse(statsResponse)
		require.NoError(err)
		require.Equal(parameter.expected, stats.TimeEstimate)
	}
}
