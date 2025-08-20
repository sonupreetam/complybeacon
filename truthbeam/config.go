package truthbeam

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
)

// Config defines configuration for the ComplianceLogProcessor.
type Config struct {
	ClientConfig confighttp.ClientConfig `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if cfg.ClientConfig.Endpoint == "" {
		return errors.New("endpoint must be specified")
	}
	return nil
}
