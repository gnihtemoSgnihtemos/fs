package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	Database    string
	Concurrency int
	Sites       []Site
	Default     Site
}

type Site struct {
	Name           string
	Address        string
	Username       string
	Password       string
	Root           string
	TLS            bool
	ConnectTimeout time.Duration
	Ignore         []string
	IgnoreSymlinks bool
}

func readConfig(r io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Config{}, err
	}
	// Unmarshal config and replace every path with the default one
	var defaults Config
	if err := json.Unmarshal(data, &defaults); err != nil {
		return Config{}, err
	}
	for i, _ := range defaults.Sites {
		defaults.Sites[i] = defaults.Default
	}
	// Unmarshal config again, letting individual paths override the defaults
	cfg := defaults
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Concurrency < 1 {
		return fmt.Errorf("concurrency must be >= 1")
	}
	if len(c.Database) == 0 {
		return fmt.Errorf("path to database must be set")
	}
	return nil
}

func ReadConfig(name string) (Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	cfg, err := readConfig(f)
	if err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
