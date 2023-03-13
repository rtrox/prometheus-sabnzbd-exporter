package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var VALID_CONFIG = Config{
	BaseURL:          "https://this.is.a.valid.url",
	ApiKey:           "acbdef0123456789acbdef0123456789",
	ListenPort:       "8080",
	LogLevel:         "info",
	GoCollector:      false,
	ProcessCollector: false,
}

func TestValidate(t *testing.T) {
	type parameter struct {
		name    string
		cfg     Config
		wantErr bool
	}

	hostPortConfig := VALID_CONFIG
	hostPortConfig.BaseURL = "localhost:8080"

	urlPortConfig := VALID_CONFIG
	urlPortConfig.BaseURL = "http://localhost:8080"

	missingBaseURLConfig := VALID_CONFIG
	missingBaseURLConfig.BaseURL = ""

	badBaseURLConfig := VALID_CONFIG
	badBaseURLConfig.BaseURL = "this is not a url"

	missingApiKeyConfig := VALID_CONFIG
	missingApiKeyConfig.ApiKey = ""

	missingPortConfig := VALID_CONFIG
	missingPortConfig.ListenPort = ""

	alphaPortConfig := VALID_CONFIG
	alphaPortConfig.ListenPort = "abc"

	badLogLevelConfig := VALID_CONFIG
	badLogLevelConfig.LogLevel = "bad"

	parameters := []parameter{
		{
			name:    "valid config - url",
			cfg:     VALID_CONFIG,
			wantErr: false,
		},
		{
			name:    "valid config - host:port",
			cfg:     hostPortConfig,
			wantErr: false,
		},
		{
			name:    "valid config - url:port",
			cfg:     urlPortConfig,
			wantErr: false,
		},
		{
			name:    "missing base url",
			cfg:     missingBaseURLConfig,
			wantErr: true,
		},
		{
			name:    "bad base url",
			cfg:     badBaseURLConfig,
			wantErr: true,
		},
		{
			name:    "missing api key",
			cfg:     missingApiKeyConfig,
			wantErr: true,
		},
		{
			name:    "missing listen port",
			cfg:     missingPortConfig,
			wantErr: true,
		},
		{
			name:    "alpha listen port",
			cfg:     alphaPortConfig,
			wantErr: true,
		},
		{
			name:    "bad log level",
			cfg:     badLogLevelConfig,
			wantErr: true,
		},
	}

	require := require.New(t)

	for _, tt := range parameters {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.NotNil(err)
			} else {
				require.Nil(err)
			}
		})
	}
}

func TestLoadConfig_Flags(t *testing.T) {
	parameters := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "defaults",
			args: []string{"--base_url", "http://localhost:8080", "--api_key", "abc123"},
			expected: Config{
				BaseURL:          "http://localhost:8080",
				ApiKey:           "abc123",
				ListenPort:       "8080",
				LogLevel:         "info",
				GoCollector:      false,
				ProcessCollector: false,
			},
		},
		{
			name: "all options",
			args: []string{
				"--base_url", "http://localhost:8080",
				"--api_key", "abc123",
				"--listen_port", "8081",
				"--log_level", "debug",
				"--go_collector", "true",
				"--process_collector", "true",
			},
			expected: Config{
				BaseURL:          "http://localhost:8080",
				ApiKey:           "abc123",
				ListenPort:       "8081",
				LogLevel:         "debug",
				GoCollector:      true,
				ProcessCollector: true,
			},
		},
	}

	require := require.New(t)

	for _, tt := range parameters {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig("testApp", tt.args)
			require.Nil(err)
			require.EqualValues(tt.expected, *cfg)
		})
	}
}

func TestLoadConfig_Env(t *testing.T) {
	parameters := []struct {
		name     string
		env      map[string]string
		expected Config
	}{
		{
			name: "defaults",
			env: map[string]string{
				"SABNZBD_BASE_URL": "http://localhost:8080",
				"SABNZBD_API_KEY":  "abc123",
			},
			expected: Config{
				BaseURL:          "http://localhost:8080",
				ApiKey:           "abc123",
				ListenPort:       "8080",
				LogLevel:         "info",
				GoCollector:      false,
				ProcessCollector: false,
			},
		},
		{
			name: "all options",
			env: map[string]string{
				"SABNZBD_BASE_URL":          "http://localhost:8080",
				"SABNZBD_API_KEY":           "abc123",
				"SABNZBD_LISTEN_PORT":       "8081",
				"SABNZBD_LOG_LEVEL":         "debug",
				"SABNZBD_GO_COLLECTOR":      "true",
				"SABNZBD_PROCESS_COLLECTOR": "true",
			},
			expected: Config{
				BaseURL:          "http://localhost:8080",
				ApiKey:           "abc123",
				ListenPort:       "8081",
				LogLevel:         "debug",
				GoCollector:      true,
				ProcessCollector: true,
			},
		},
	}

	require := require.New(t)

	for _, tt := range parameters {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			cfg, err := LoadConfig("testApp", []string{})
			require.Nil(err)
			require.EqualValues(tt.expected, *cfg)
		})
	}
}

func TestLoadConfig_File(t *testing.T) {
	parameters := []struct {
		name     string
		file     string
		expected Config
	}{
		{
			name: "defaults",
			file: "test_fixtures/defaults.yaml",
			expected: Config{
				BaseURL:          "http://localhost:8080",
				ApiKey:           "abc123",
				ListenPort:       "8080",
				LogLevel:         "info",
				GoCollector:      false,
				ProcessCollector: false,
			},
		},
		{
			name: "all options",
			file: "test_fixtures/all_options.yaml",
			expected: Config{
				BaseURL:          "http://localhost:8080",
				ApiKey:           "abc123",
				ListenPort:       "8081",
				LogLevel:         "debug",
				GoCollector:      true,
				ProcessCollector: true,
			},
		},
	}

	require := require.New(t)

	for _, tt := range parameters {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig("testApp", []string{"--config", tt.file})
			require.Nil(err)

			require.EqualValues(tt.expected, *cfg)
		})
	}
}
