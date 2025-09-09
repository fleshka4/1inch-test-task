package config

import (
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config holds application configuration loaded from file.
type Config struct {
	RPCURL            string        `yaml:"rpc_url"`
	ListenAddr        string        `yaml:"listen_addr"`
	GraceTimeout      time.Duration `yaml:"shutdown_timeout"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	CallTimeout       time.Duration `yaml:"call_timeout"`
}

// Load reads the config from a YAML file path.
// Fails if config is invalid or file is missing.
func Load(path string) (*Config, error) {
	//nolint:gosec
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "os.Open")
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close config file: file.Close: %v", err)
		}
	}()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, errors.Wrap(err, "decoder.Decode")
	}

	if cfg.RPCURL == "" {
		return nil, errors.New("rpc_url is required")
	}

	cfg.applyDefaults()

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	const (
		defaultTimeout = 5 * time.Second
		listenAddr     = ":1337"
	)

	if c.ListenAddr == "" {
		c.ListenAddr = listenAddr
	}
	if c.GraceTimeout <= 0 {
		c.GraceTimeout = defaultTimeout
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = defaultTimeout
	}
	if c.ReadHeaderTimeout <= 0 {
		c.ReadHeaderTimeout = defaultTimeout
	}
	if c.CallTimeout <= 0 {
		c.CallTimeout = defaultTimeout
	}
}
