package truthbeam

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/complytime/complybeacon/truthbeam/internal/client"
)

// Config defines configuration for the truthbeam processor.
type Config struct {
	ClientConfig  confighttp.ClientConfig `mapstructure:",squash"`        // squash ensures fields are correctly decoded in embedded struct.
	CacheTTL      time.Duration           `mapstructure:"cache_ttl"`      // Cache TTL for compliance metadata
	CacheCapacity int                     `mapstructure:"cache_capacity"` // Cache capacity in number of entries (0 = use default from client.DefaultCacheCapacity)
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if cfg.ClientConfig.Endpoint == "" {
		return errors.New("endpoint must be specified")
	}
	// Normalize cache TTL: 0 means use default (24 hours for compliance metadata)
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = client.DefaultCacheTTL
	}
	// Set default cache capacity if not specified
	if cfg.CacheCapacity == 0 {
		cfg.CacheCapacity = client.DefaultCacheCapacity
	}
	if cfg.CacheCapacity < 0 {
		return errors.New("cache_capacity must be non-negative")
	}

	return nil
}
