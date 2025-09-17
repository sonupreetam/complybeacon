package factory

import (
	"github.com/complytime/complybeacon/compass/mapper"
	"github.com/complytime/complybeacon/compass/mapper/plugins/basic"
)

func MapperByID(_ mapper.ID) mapper.Mapper {
	return basic.NewBasicMapper()
}
