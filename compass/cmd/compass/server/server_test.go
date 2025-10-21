package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScopeFromCatalogPath(t *testing.T) {
	t.Run("returns error for non-existent file", func(t *testing.T) {
		// Test creating a scope from a non-existent file (no scope created)
		scope, err := NewScopeFromCatalogPath("/non/existent/catalog-file.yaml")

		assert.Error(t, err)
		assert.Nil(t, scope)
	})
}
