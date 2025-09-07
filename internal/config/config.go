package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration loaded from file.
type Config struct {
	RPCURL            string        `yaml:"rpc_url"`
	ListenAddr        string        `yaml:"listen_addr"`
	GraceTimeout      time.Duration `yaml:"shutdown_timeout"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
}

// Load reads the config from a YAML file path.
// Fails fatally if config is invalid or file is missing.
func Load(path string) Config {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open config file: os.Open: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("failed to close config file: f.Close: %v", err)
		}
	}(f)

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		log.Fatalf("failed to parse config file: decoder.Decode: %v", err)
	}

	// Fallbacks
	const defaultTimeout = 5 * time.Second
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":1337"
	}
	if cfg.GraceTimeout == 0 {
		cfg.GraceTimeout = defaultTimeout
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = defaultTimeout
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = defaultTimeout
	}

	if cfg.RPCURL == "" {
		log.Fatalf("rpc_url is required in config")
	}

	return cfg
}
