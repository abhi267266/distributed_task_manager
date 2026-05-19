package common

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds the configuration for the central server.
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Address      string `yaml:"address"`
	SharedSecret string `yaml:"shared_secret"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
}

// AgentConfig holds the configuration for the distributed agent.
type AgentConfig struct {
	ServerURL      string `yaml:"server_url"`
	Interval       int    `yaml:"interval"` // Heartbeat interval in seconds
	SharedSecret   string `yaml:"shared_secret"`
	ServerCertFile string `yaml:"server_cert_file"`
}

// LoadConfig loads a YAML configuration file into the provided struct.
func LoadConfig(path string, config interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	// Check environment variable for secret, fallback to config
	envSecret := os.Getenv("BOTNET_SECRET")
	if envSecret != "" {
		if serverCfg, ok := config.(*ServerConfig); ok {
			serverCfg.SharedSecret = envSecret
		} else if agentCfg, ok := config.(*AgentConfig); ok {
			agentCfg.SharedSecret = envSecret
		}
	}

	return nil
}
