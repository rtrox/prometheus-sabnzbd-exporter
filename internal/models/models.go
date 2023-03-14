package models

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Status int

const (
	KB = 1024
	MB = 1024 * KB
)

const (
	UNKNOWN Status = iota
	IDLE
	PAUSED
	DOWNLOADING
)

func (s Status) Float64() float64 {
	return float64(s)
}
func (s Status) String() string {
	switch s {
	case IDLE:
		return "Idle"
	case PAUSED:
		return "Paused"
	case DOWNLOADING:
		return "Downloading"
	default:
		return "Unknown"
	}
}

func StatusFromString(s string) Status {
	switch s {
	case "Idle":
		return IDLE
	case "Paused":
		return PAUSED
	case "Downloading":
		return DOWNLOADING
	default:
		return UNKNOWN
	}
}

type ServerStat struct {
	Total           int    // Total Data Downloaded in bytes
	ArticlesTried   int    // Number of Articles Tried
	ArticlesSuccess int    // Number of Articles Successfully Downloaded
	DayParsed       string // Last Date Parsed
}

type ServerStats struct {
	Total   int // Total Data Downloaded in bytes
	Servers map[string]ServerStat
}

func NewServerStatsFromResponse(response ServerStatsResponse) *ServerStats {
	ret := &ServerStats{
		Total:   response.Total,
		Servers: make(map[string]ServerStat),
	}

	for name, stats := range response.Servers {
		d, tried := latestStat(stats.ArticlesTried)
		_, success := latestStat(stats.ArticlesSuccess)
		ret.Servers[name] = struct {
			Total           int
			ArticlesTried   int
			ArticlesSuccess int
			DayParsed       string
		}{
			Total:           stats.Total,
			ArticlesTried:   tried,
			ArticlesSuccess: success,
			DayParsed:       d,
		}
	}

	return ret
}

type QueueStats struct {
	Version                    string        // Sabnzbd Version
	Paused                     bool          // Is the sabnzbd queue globally paused?
	PausedAll                  bool          // Paused All actions which causes disk activity
	PauseDuration              time.Duration // Duration sabnzbd will remain paused
	DownloadDirDiskspaceUsed   float64       // Download Directory Used in bytes
	DownloadDirDiskspaceTotal  float64       // Download Directory Total in bytes
	CompletedDirDiskspaceUsed  float64       // Completed Directory Used in bytes
	CompletedDirDiskspaceTotal float64       // Completed Directory Total in bytes
	SpeedLimit                 float64       // The Speed Limit set as a percentage of configured line speed
	SpeedLimitAbs              float64       // The Speed Limit set in B/s
	HaveWarnings               float64       // Number of Warnings present
	Quota                      float64       // Total Quota configured Bytes
	HaveQuota                  bool          // Is a Periodic Quota set for Sabnzbd?
	RemainingQuota             float64       // Quota Remaining Bytes
	CacheArt                   float64       // Number of Articles in Cache
	CacheSize                  float64       // Size of Cache in bytes
	Speed                      float64       // Float String representing bps
	RemainingSize              float64       // Bytes left to download in queue
	Size                       float64       // total bytes represented by queue
	ItemsInQueue               float64       // Total number of items in queue
	Status                     Status        // Status of sabnzbd (1 = Idle, 2 = Paused, 3 = Downloading)
	TimeEstimate               time.Duration // Estimated time remaining to download queue
}

func NewQueueStatsFromResponse(response QueueResponse) (QueueStats, error) {
	var err error

	queue := response.Queue
	pauseDuration, err := parseDuration(queue.PauseInt, err)
	downloadDirDiskspaceUsed, err := parseFloat(queue.Diskspace1, err)
	downloadDirDiskspaceTotal, err := parseFloat(queue.DiskspaceTotal1, err)
	completedDirDiskspaceUsed, err := parseFloat(queue.Diskspace2, err)
	completedDirDiskspaceTotal, err := parseFloat(queue.DiskspaceTotal2, err)
	leftQuota, err := parseSize(queue.LeftQuota, err)
	cacheArt, err := parseSize(queue.CacheArt, err)
	cacheSize, err := parseSize(queue.CacheSize, err)
	speed, err := parseFloat(queue.KBPerSec, err)
	remainingSize, err := parseFloat(queue.MBLeft, err)
	size, err := parseFloat(queue.MB, err)
	quota, err := parseSize(queue.Quota, err)
	speedLimit, err := parseSize(queue.Speedlimit, err)
	speedLimitAbs, err := parseSize(queue.SpeedlimitAbs, err)
	haveWarnings, err := parseFloat(queue.HaveWarnings, err)
	timeLeft, err := parseDuration(queue.TimeLeft, err)

	if err != nil {
		return QueueStats{}, fmt.Errorf("Error parsing queue stats: %s", err)
	}

	return QueueStats{
		Version:                    queue.Version,
		Paused:                     queue.Paused,
		PausedAll:                  queue.PausedAll,
		PauseDuration:              pauseDuration,
		DownloadDirDiskspaceUsed:   downloadDirDiskspaceUsed * MB,
		DownloadDirDiskspaceTotal:  downloadDirDiskspaceTotal * MB,
		CompletedDirDiskspaceUsed:  completedDirDiskspaceUsed * MB,
		CompletedDirDiskspaceTotal: completedDirDiskspaceTotal * MB,
		SpeedLimit:                 speedLimit,
		SpeedLimitAbs:              speedLimitAbs,
		HaveWarnings:               haveWarnings,
		Quota:                      quota,
		HaveQuota:                  queue.HaveQuota,
		RemainingQuota:             leftQuota,
		CacheArt:                   cacheArt,
		CacheSize:                  cacheSize,
		Speed:                      speed * KB,
		RemainingSize:              remainingSize * MB,
		Size:                       size * MB,
		ItemsInQueue:               float64(queue.NoofSlotsTotal),
		Status:                     StatusFromString(queue.Status),
		TimeEstimate:               timeLeft,
	}, nil
}

// latestStat gets the most recent date's value from a map of dates to values
func latestStat(m map[string]int) (string, int) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	key := keys[len(keys)-1]

	return key, m[key]
}

// parseFloat is a monad version of strconv.ParseFloat
func parseFloat(f string, prevErr error) (float64, error) {
	if prevErr != nil {
		return 0, prevErr
	}

	if f == "" {
		return 0, nil
	}

	ret, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return 0, err
	}

	return ret, nil
}

// parseSize is a monad which parses a size string in the format of "123.45 KB" or "123.45"
func parseSize(sz string, prevErr error) (float64, error) {
	if prevErr != nil {
		return 0, prevErr
	}

	fields := strings.Fields(strings.TrimSpace(sz))
	if len(fields) == 0 {
		return 0, nil
	}

	if len(fields) > 2 {
		return 0, fmt.Errorf("Invalid size: %s", sz)
	}

	ret, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	if len(fields) == 1 {
		return ret, nil
	}

	switch fields[1] {
	case "B":
		return ret, nil
	case "KB", "K":
		return ret * 1024, nil
	case "MB", "M":
		return ret * 1024 * 1024, nil
	case "GB", "G":
		return ret * 1024 * 1024 * 1024, nil
	case "TB", "T":
		return ret * 1024 * 1024 * 1024 * 1024, nil
	case "PB", "P":
		return ret * 1024 * 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("Invalid size suffix: %s", sz)
	}
}

// parseDuration is a monad which parses a duration string in the format of "HH:MM:SS" or "MM:SS"
func parseDuration(s string, prevErr error) (time.Duration, error) {
	if prevErr != nil {
		return 0, prevErr
	}

	if s == "" {
		return 0, nil
	}

	fields := strings.Split(strings.TrimSpace(s), ":")
	if len(fields) < 1 || len(fields) > 4 {
		return 0, fmt.Errorf("Invalid duration: %s", s)
	}

	intFields := make([]int, len(fields))

	for i, f := range fields {
		var err error
		// Reverse the order of the fields
		intFields[len(intFields)-1-i], err = strconv.Atoi(f)
		if err != nil {
			return 0, fmt.Errorf("Invalid integer in duration: %s: %w", f, err)
		}
	}

	ret := time.Duration(intFields[0]) * time.Second

	fieldCount := len(intFields)
	if fieldCount > 1 {
		ret += time.Duration(intFields[1]) * time.Minute
	}

	if fieldCount > 2 {
		ret += time.Duration(intFields[2]) * time.Hour
	}

	if fieldCount > 3 {
		ret += time.Duration(intFields[3]) * 24 * time.Hour
	}

	return ret, nil
}
