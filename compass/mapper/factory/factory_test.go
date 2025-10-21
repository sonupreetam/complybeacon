package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/complytime/complybeacon/compass/mapper/plugins/basic"
)

func TestMapperByID(t *testing.T) {
	t.Run("returns basic mapper for any ID", func(t *testing.T) {
		// Get the basic mapper for any ID
		result := MapperByID("test-id")

		assert.NotNil(t, result)
		assert.IsType(t, &basic.Mapper{}, result)
		assert.Equal(t, basic.ID, result.PluginName())
	})

	t.Run("returns basic mapper for empty ID", func(t *testing.T) {
		// Get the basic mapper for empty ID
		result := MapperByID("")

		assert.NotNil(t, result)
		assert.IsType(t, &basic.Mapper{}, result)
	})

	t.Run("returns basic mapper for special characters", func(t *testing.T) {
		// Get the basic mapper for special characters
		result := MapperByID("mapper-v1.0")

		assert.NotNil(t, result)
		assert.IsType(t, &basic.Mapper{}, result)
	})
}
