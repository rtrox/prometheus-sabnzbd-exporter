package config

import (
	"fmt"
	"os"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

var ENV_PREFIX = "SABNZBD_"

type Config struct {
	BaseURL          string `koanf:"base_url"`
	ApiKey           string `koanf:"api_key"`
	ListenPort       string `koanf:"listen_port"`
	LogLevel         string `koanf:"log_level"`
	GoCollector      bool   `koanf:"go-collector"`
	ProcessCollector bool   `koanf:"process-collector"`
}

func LoadConfig(appName string, args []string) (*Config, error) {
	k := koanf.New(".")
	f := flag.NewFlagSet(appName, flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.StringSlice("config", []string{}, "path to one or more .yaml config files")
	f.String("log_level", "info", "log level (debug, info, warn, error)")
	f.Bool("go_collector", false, "enables go stats exporter")
	f.Bool("process_collector", false, "enables process stats exporter")
	f.String("listen_port", "8080", "port to listen on")
	f.String("base_url", "", "base url of sabnzbd")
	f.String("api_key", "", "api key of sabnzbd")
	f.Parse(os.Args[1:])

	k.Load(confmap.Provider(map[string]interface{}{
		"log_level":         "info",
		"listen_port":       "8080",
		"go_collector":      false,
		"process_collector": false,
	}, "."), nil)
	cFiles, _ := f.GetStringSlice("config")
	for _, c := range cFiles {
		if err := k.Load(file.Provider(c), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("Error loading config file (%s): %w", c, err)
		}
	}

	k.Load(env.Provider("SABNZBD_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "SABNZBD_"))
	}), nil)

	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		return nil, fmt.Errorf("Error loading flags: %w", err)
	}

	var out Config
	err := k.Unmarshal("", &out)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling config: %w", err)
	}
	return &out, nil

}

func (c *Config) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.BaseURL, validation.Required),
		validation.Field(&c.ApiKey, validation.Required),
		validation.Field(&c.ListenPort, validation.Required, is.Port),
		validation.Field(&c.LogLevel, validation.Required, validation.In("debug", "info", "warn", "error")),
	)
}
