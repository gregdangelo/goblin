package config

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/pelletier/go-toml/v2"
)

type ConfigStore struct {
	current atomic.Pointer[ChaosConfig]
}

func NewConfigStore(cfg ChaosConfig) *ConfigStore {
	store := &ConfigStore{}
	store.current.Store(&cfg)
	return store
}

// should I allow custom path? probably but not right now
func LoadFromFile() (*ConfigStore, error) {
	// load the chaos config from path then run NewConfigStore
	file, err := os.OpenInRoot(".", "goblin.toml")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var cfg ChaosConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return NewConfigStore(cfg), nil
}

func (s *ConfigStore) Get() ChaosConfig {
	return *s.current.Load()
}

func (s *ConfigStore) Set(cfg ChaosConfig) {
	s.current.Store(&cfg)
}

func (s *ConfigStore) Save(path string) error {
	data, err := toml.Marshal(s.current.Load())
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
