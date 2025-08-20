package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/complytime/complybeacon/compass/api"
	"github.com/complytime/complybeacon/compass/transformer"
	"github.com/complytime/complybeacon/compass/transformer/plugins/basic"
)

// Service struct to hold dependencies if needed
type Service struct {
	transformers transformer.Set
	scope        Scope
}

// NewService initializes a new Service instance.
func NewService(transformers transformer.Set, scope Scope) *Service {
	return &Service{
		transformers: transformers,
		scope:        scope,
	}
}

// PostV1Enrich handles the POST /v1/enrich endpoint.
// It's a handler function for Gin.
func (s *Service) PostV1Enrich(c *gin.Context) {
	var req api.EnrichmentRequest
	err := c.Bind(&req)
	if err != nil {
		sendCompassError(c, http.StatusBadRequest, "Invalid format for enrichment")
		return
	}

	transformationPlugin, ok := s.transformers[transformer.ID(req.Evidence.Source)]
	if !ok {
		// Use fallback
		transformationPlugin = basic.NewBasicTransformer()
	}
	enrichedResponse := Enrich(req.Evidence, transformationPlugin, s.scope)

	c.JSON(http.StatusOK, enrichedResponse)
}

// sendCompassError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendCompassError(c *gin.Context, code int32, message string) {
	compassErr := api.Error{
		Code:    code,
		Message: message,
	}
	c.JSON(int(code), compassErr)
}
