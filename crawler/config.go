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
	Skip           bool
	ConnectTimeout string
	connectTimeout time.Duration
	ReadTimeout    string
	readTimeout    time.Duration
	Ignore         []string
	IgnoreSymlinks bool
}

func readConfig(r io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Config{}, err
	}
	// Unmarshal config and replace every site with the default one
	var defaults Config
	if err := json.Unmarshal(data, &defaults); err != nil {
		return Config{}, err
	}
	for i, _ := range defaults.Sites {
		defaults.Sites[i] = defaults.Default
		defaults.Sites[i].Ignore = make([]string, len(defaults.Default.Ignore))
		copy(defaults.Sites[i].Ignore, defaults.Default.Ignore)
	}
	// Unmarshal config again, letting individual sites override the defaults
	cfg := defaults
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Concurrency < 1 {
		return fmt.Errorf("concurrency must be >= 1")
	}
	if len(c.Database) == 0 {
		return fmt.Errorf("path to database must be set")
	}
	for i, site := range c.Sites {
		{
			d, err := time.ParseDuration(site.ConnectTimeout)
			if err != nil {
				return err
			}
			c.Sites[i].connectTimeout = d
		}
		{
			d, err := time.ParseDuration(site.ReadTimeout)
			if err != nil {
				return err
			}
			c.Sites[i].readTimeout = d
		}
	}
	return nil
}

func (c *Config) JSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func ReadConfig(name string) (Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	return readConfig(f)
}
