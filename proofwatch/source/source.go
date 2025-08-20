package source

import (
	"context"
	"errors"
	"fmt"

	"github.com/complytime/complybeacon/proofwatch/evidence"
	"github.com/complytime/complybeacon/proofwatch/source/push"
)

type Source interface {
	Name() string
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Config struct {
	instrument evidence.InstrumentationFn
	observer   *evidence.EvidenceObserver
	PushConfig *push.Config `yaml:"push,omitempty"`
}

func (c *Config) SetupObserver(observer *evidence.EvidenceObserver) {
	c.observer = observer
	c.instrument = evidence.NewEmitter(observer)
}

func Deploy(ctx context.Context, cfg Config) ([]Source, error) {
	var sources []Source
	switch {
	case cfg.PushConfig != nil:
		source := push.NewSource(cfg.PushConfig, cfg.instrument)
		sources = append(sources, source)
	default:
		return sources, fmt.Errorf("no valid config defined")
	}

	var errs []error
	for _, source := range sources {
		if err := source.Run(ctx); err != nil {
			errs = append(errs, fmt.Errorf("source %s: %w", source.Name(), err))
		}
	}

	return sources, errors.Join(errs...)
}

func Shutdown(ctx context.Context, sources []Source) error {
	var errs []error
	for _, source := range sources {
		if err := source.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("source %s: %w", source.Name(), err))
		}
	}
	return errors.Join(errs...)
}
