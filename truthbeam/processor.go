package truthbeam

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"

	"github.com/complytime/complybeacon/truthbeam/internal/client"
)

type truthBeamProcessor struct {
	telemetry component.TelemetrySettings
	config    *Config

	log *zap.Logger

	client *client.Client

	// TODO: Cache results by policy id
}

func newTruthBeamProcessor(conf component.Config, set processor.Settings) (*truthBeamProcessor, error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return nil, errors.New("invalid configuration provided")
	}

	// TODO: Apply additional options from client config including
	// mTLS settings
	cl, err := client.NewClient(cfg.ClientConfig.Endpoint)
	if err != nil {
		return nil, err
	}

	return &truthBeamProcessor{
		config:    cfg,
		telemetry: set.TelemetrySettings,
		log:       set.Logger,
		client:    cl,
	}, nil
}

func (t *truthBeamProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	rl := ld.ResourceLogs()
	for i := 0; i < rl.Len(); i++ {
		rs := rl.At(i)
		ilss := rs.ScopeLogs()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			logs := ils.LogRecords()
			resource := rs.Resource()
			for k := 0; k < logs.Len(); k++ {
				logRecord := logs.At(k)
				err := client.ApplyAttributes(ctx, t.client, t.config.ClientConfig.Endpoint, resource, logRecord)
				if err != nil {
					// We don't want to return an error here to ensure the evidence
					// is not dropped. It will just be uncategorized.
					t.log.Error(err.Error())
				}
			}
		}
	}
	return ld, nil
}
