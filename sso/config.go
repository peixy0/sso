package sso

import (
	"encoding/json"
	"fmt"
	"os"
)

type Service struct {
	Name        string `json:"name"`
	CallbackURL string `json:"callbackUrl"`
	Key         string `json:"key"`
}

type ServiceRegistry struct {
	services map[string]Service
}

func LoadServiceRegistry(path string) (*ServiceRegistry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open service config: %w", err)
	}
	defer file.Close()

	var configs []Service
	if err := json.NewDecoder(file).Decode(&configs); err != nil {
		return nil, fmt.Errorf("failed to decode service config: %w", err)
	}

	registry := &ServiceRegistry{
		services: make(map[string]Service),
	}
	for _, config := range configs {
		registry.services[config.Name] = config
	}

	return registry, nil
}

func (r *ServiceRegistry) Get(name string) (Service, bool) {
	config, ok := r.services[name]
	return config, ok
}
