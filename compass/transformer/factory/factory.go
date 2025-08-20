package factory

import (
	"github.com/complytime/complybeacon/compass/transformer"
	"github.com/complytime/complybeacon/compass/transformer/plugins/basic"
)

func TransformerByID(_ transformer.ID) transformer.Transformer {
	return basic.NewBasicTransformer()
}
