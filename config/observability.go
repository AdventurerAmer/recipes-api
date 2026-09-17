package config

import (
	"time"

	"github.com/AdventurerAmer/recipes-api/logging"
)

type Observability struct {
	Logging      Logging      `koanf:"logging"`
	Tracing      Tracing      `koanf:"tracing"`
	Metrics      Metrics      `koanf:"metrics"`
	HealthChecks HealthChecks `koanf:"healthChecks"`
}

type Logging struct {
	Level     logging.Level  `koanf:"level" validate:"oneof=debug info warn error"`
	Format    logging.Format `koanf:"format" validate:"oneof=json text"`
	AddSource *bool          `koanf:"addSource"`
}

type Tracing struct {
	Enabled    bool    `koanf:"enabled"`
	Endpoint   string  `koanf:"endpoint" validate:"required,url"`
	SampleRate float64 `koanf:"sampleRate" validate:"required,min=0.0,max=1.0"`
}

type Metrics struct {
	Enabled  bool   `koanf:"enabled"`
	Endpoint string `koanf:"endpoint" validate:"required,url"`
	Runtime  bool   `koanf:"runtime"`
	Requests bool   `koanf:"requests"`
}

type HealthChecks struct {
	Enabled  bool          `koanf:"enabled"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
	Checks   []string      `koanf:"checks"`
}
