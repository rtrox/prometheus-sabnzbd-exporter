package models

// ServerStatsResponse is the response from the sabnzbd serverstats endpoint
type ServerStatsResponse struct {
	Total   int                           `json:"total"`
	Servers map[string]ServerStatResponse `json:"servers"`
}

type ServerStatResponse struct {
	Total           int            `json:"total"`            // Total Data Downloaded in bytes
	ArticlesTried   map[string]int `json:"articles_tried"`   // Number of Articles Tried (YYYY-MM-DD -> count)
	ArticlesSuccess map[string]int `json:"articles_success"` // Number of Articles Successfully Downloaded (YYYY-MM-DD -> count)
}

// QueueResponse is the response from the sabnzbd queue endpoint
// Paused vs PausedAll -- as best I can tell, Paused is
// "pause the queue but finish anything in flight"
// PausedAll is "hard pause, including pausing in progress downloads"
type QueueResponse struct {
	Queue QueueResponseQueue `json:"queue"`
}

type QueueResponseQueue struct {
	Version         string `json:"version"`         // version of Sabnzbd running
	Paused          bool   `json:"paused"`          // Is the sabnzbd queue globally paused?
	PauseInt        string `json:"pause_int"`       // returns minutes:seconds until sabnzbd is unpaused (minutes are unpadded)
	PausedAll       bool   `json:"paused_all"`      // Paused All actions which causes disk activity
	Diskspace1      string `json:"diskspace1"`      // Download Directory Used (float, MB)
	Diskspace2      string `json:"diskspace2"`      // Completed Directory Used (float, MB)
	DiskspaceTotal1 string `json:"diskspacetotal1"` // Download Directory Total (float, MB)
	DiskspaceTotal2 string `json:"diskspacetotal2"` // Completed Directory Total (float, MB)
	Speedlimit      string `json:"speedlimit"`      // The Speed Limit set as a percentage of configured line speed
	SpeedlimitAbs   string `json:"speedlimitabs"`   // The Speed Limit set in B/s
	HaveWarnings    string `json:"have_warnings"`   // Number of Warnings present
	Quota           string `json:"quota"`           // Total Quota configured (normalized to K/M/G/T/P)
	HaveQuota       bool   `json:"have_quota"`      // Is a Periodic Quota set for Sabnzbd?
	LeftQuota       string `json:"left_quota"`      // Quota Remaining (normalized to K/M/G/T/P)
	CacheArt        string `json:"cache_art"`       // Number of Articles in Cache
	CacheSize       string `json:"cache_size"`      // Size of Cache in bytes (normalized to "B/MB/GB/TB/PB")
	KBPerSec        string `json:"kbpersec"`        // Float String representing Kbps
	MBLeft          string `json:"mbleft"`          // Megabytes left to download in queue
	MB              string `json:"mb"`              // total megabytes represented by queue
	NoofSlotsTotal  int    `json:"noofslots_total"` // Total number of items in queue
	Status          string `json:"status"`          // Status of sabnzbd (Paused, Idle, Downloading)
	TimeLeft        string `json:"timeleft"`        // Estimated time to download all items in queue (HH:MM:SS)

	// Speed           string `json:"speed"`        // Float String normalized to B/K/M/G/T/P
	// SizeLeft        string `json:"sizeleft"`     // Bytes left to download in queue (normalized to "B/KB/MB/GB/TB/PB")
	// Size            string `json:"size"`         // total bytes represented by queue (normalized to "B/KB/MB/GB/TB/PB")
	// Start           int    `json:"start"`        // Index of first item in queue (0 based)
	//Limit           int    `json:"limit"`         // Number of items to return in response
	//Finish          int    `json:"finish"`        // Index of last item in queue (0 based)
	// NoofSlots int `json:"noofslots"` // Number of slots in api response (may be less than total if limit is set)
}
