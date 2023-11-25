// Date: 2023/11/25
// Created By ybenel
package app

import (
	"encoding/json"
	"os"
)

// Config settings for main App.
type Config struct {
	Library []*PathConfig  `json:"library"`
	Server  *ServerConfig  `json:"server"`
	Feed    *FeedConfig    `json:"feed"`
	Tor     *TorConfig     `json:"tor,omitempty"`
	Stremio *StremioConfig `json:"stremio"`
	Logging bool           `json:"logging"`
}

// PathConfig settings for media library path.
type PathConfig struct {
	Path    string `json:"path"`
	Prefix  string `json:"prefix"`
	Private bool   `json:"private"`
}

// ServerConfig settings for App Server.
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// FeedConfig settings for App Feed.
type FeedConfig struct {
	ExternalURL string `json:"external_url"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Author      struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
	Copyright string `json:"copyright"`
}

// TorConfig stores tor configuration.
type TorConfig struct {
	Enable     bool                 `json:"enable"`
	Controller *TorControllerConfig `json:"controller"`
}

// Stremio Config
type StremioConfig struct {
	StreamUrl string `json:"host"`
}

// TorControllerConfig stores tor controller configuration.
type TorControllerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password,omitempty"`
}

// DefaultConfig returns Config initialized with default values.
func DefaultConfig() *Config {
	return &Config{
		Library: []*PathConfig{
			{
				Path:    "videos",
				Prefix:  "",
				Private: false,
			},
		},
		Server: &ServerConfig{
			Host: "127.0.0.1",
			Port: 0,
		},
		Stremio: &StremioConfig{
			StreamUrl: "http://127.0.0.1:8080",
		},
		Feed: &FeedConfig{
			ExternalURL: "http://localhost",
		},
		Tor: &TorConfig{
			Enable: false,
			Controller: &TorControllerConfig{
				Host: "127.0.0.1",
				Port: 9051,
			},
		},
		Logging: true,
	}
}

// ReadFile reads a JSON file into Config.
func (c *Config) ReadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	return d.Decode(c)
}
