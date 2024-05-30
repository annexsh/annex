package server

import (
	"fmt"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type Env string

const (
	EnvLocal   = "local"
	EnvNonProd = "nonprod"
	EnvProd    = "prod"
)

func (e Env) validate() error {
	switch e {
	case EnvLocal, EnvNonProd, EnvProd:
		return nil
	default:
		return fmt.Errorf("invalid env '%s'", e)
	}
}

type Config struct {
	Env      Env      `yaml:"env"`
	Port     int      `yaml:"port" required:"true"`
	Temporal Temporal `yaml:"temporal"`
	Postgres Postgres `yaml:"postgres"`
	InMemory bool     `yaml:"inMemory"` // temporary option during initial development phase (overrides Postgres when set)
}

func (c Config) Validate() error {
	return c.Env.validate()
}

type Temporal struct {
	LocalDev  bool   `yaml:"localDev"` // temporary option during initial development phase
	HostPort  string `yaml:"hostPort" required:"true"`
	Namespace string `yaml:"namespace" required:"true"`
}

type Postgres struct {
	SchemaVersion uint   `yaml:"schemaVersion" required:"true"`
	Host          string `required:"true"`
	Port          string `required:"true"`
	Database      string `required:"true"`
	User          string `required:"true"`
	Password      string `required:"true"`
}

func (p Postgres) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", p.User, p.Password, p.Host, p.Port, p.Database)
}

func LoadConfig(opts ...ConfigOption) (Config, error) {
	loaderCfg := aconfig.Config{
		FileDecoders: map[string]aconfig.FileDecoder{},
	}

	for _, opt := range opts {
		opt(&loaderCfg)
	}

	var cfg Config
	loader := aconfig.LoaderFor(&cfg, loaderCfg)
	if err := loader.Load(); err != nil {
		return cfg, err
	}

	return cfg, cfg.Validate()
}

type ConfigOption func(loadCfg *aconfig.Config)

func WithYAML() ConfigOption {
	return func(loaderCfg *aconfig.Config) {
		loaderCfg.FileFlag = "config-file"
		loaderCfg.FailOnFileNotFound = true
		loaderCfg.FileDecoders[".yaml"] = aconfigyaml.New()
	}
}
