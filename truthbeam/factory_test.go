package truthbeam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confighttp"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	config := factory.CreateDefaultConfig()

	require.NotNil(t, config, "Config should not be nil")

	cfg, ok := config.(*Config)
	require.True(t, ok, "Expected *Config, got %T", config)

	assert.Empty(t, cfg.ClientConfig.Endpoint, "Expected default endpoint to be empty (must be set by user)")
	assert.Equal(t, 30*time.Second, cfg.ClientConfig.Timeout, "Expected timeout 30s")
	assert.Equal(t, configcompression.Type("gzip"), cfg.ClientConfig.Compression, "Expected gzip compression")
	assert.Equal(t, 512*1024, cfg.ClientConfig.WriteBufferSize, "Expected write buffer size 512KB")
}

func TestCreateLogsProcessor(t *testing.T) {
	factory := NewFactory()
	config := factory.CreateDefaultConfig()

	assert.Equal(t, "truthbeam", factory.Type().String(), "Expected factory type 'truthbeam'")

	cfg := config.(*Config)
	err := cfg.Validate()
	assert.Error(t, err, "Expected config validation to fail for empty endpoint")
	assert.Contains(t, err.Error(), "endpoint must be specified")

	cfg.ClientConfig.Endpoint = "http://localhost:8081"
	err = cfg.Validate()
	assert.NoError(t, err, "Config validation should pass with endpoint set")
}

func TestConfigValidation(t *testing.T) {
	validConfig := getValidConfig()
	err := validConfig.Validate()
	assert.NoError(t, err, "Valid config should pass validation")

	invalidConfig := getInvalidConfig()
	err = invalidConfig.Validate()
	assert.Error(t, err, "Invalid config should fail validation")
	assert.Contains(t, err.Error(), "endpoint must be specified")
}

func TestConfigFromTestdata(t *testing.T) {
	validConfig := getValidConfig()
	defaultConfig := getDefaultConfig()
	invalidConfig := getInvalidConfig()

	require.NotNil(t, validConfig, "Valid config should not be nil")
	require.NotNil(t, defaultConfig, "Default config should not be nil")
	require.NotNil(t, invalidConfig, "Invalid config should not be nil")

	err := defaultConfig.Validate()
	assert.Error(t, err, "Default config should fail validation (empty endpoint)")
	assert.Contains(t, err.Error(), "endpoint must be specified")
}

// Helper functions for test configurations
func getValidConfig() *Config {
	return &Config{
		ClientConfig: confighttp.ClientConfig{
			Endpoint:        "http://localhost:8081",
			Timeout:         30 * time.Second,
			Compression:     "gzip",
			WriteBufferSize: 512 * 1024,
			ReadBufferSize:  0,
		},
	}
}

func getInvalidConfig() *Config {
	return &Config{
		ClientConfig: confighttp.ClientConfig{
			// Missing required endpoint
			Timeout:     30 * time.Second,
			Compression: "gzip",
		},
	}
}

func getDefaultConfig() *Config {
	clientConfig := confighttp.NewDefaultClientConfig()
	clientConfig.Timeout = 30 * time.Second
	clientConfig.Compression = "gzip"
	clientConfig.WriteBufferSize = 512 * 1024

	return &Config{
		ClientConfig: clientConfig,
	}
}
