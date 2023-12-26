package apksm

import (
	"apksm/validate"
	"encoding/json"
)

type Config struct {
	Services Services  `json:"services"`
	Settings *Settings `json:"settings"`
}

// NewConfig returns pointer to Config which is created from provided JSON data.
// Guarantees to be validated.
func NewConfig(jsonData []byte) *Config {
	config := &Config{}
	err := json.Unmarshal(jsonData, config)
	if err != nil {
		panic("error parsing json configuration data")
	}
	if err := validate.ValidateAll(config); err != nil {
		panic(err)
	}
	return config
}
