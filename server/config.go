package server

import (
	"fmt"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type AllInOneConfig struct {
	Port        int               `yaml:"port" required:"true"`
	Postgres    PostgresConfig    `yaml:"postgres"`
	Temporal    TemporalConfig    `yaml:"temporal"`
	Development DevelopmentConfig `yaml:"development"`
}

type TestServiceConfig struct {
	Port                int            `yaml:"port" required:"true"`
	Postgres            PostgresConfig `yaml:"postgres"`
	WorkflowServicePort int            `yaml:"workflowHostPort" required:"true"`
}

type EventServiceConfig struct {
	Port     int            `yaml:"port" required:"true"`
	Postgres PostgresConfig `yaml:"postgres"`
}

type WorkflowProxyServiceConfig struct {
	Port            int            `yaml:"port" required:"true"`
	Temporal        TemporalConfig `yaml:"temporal"`
	TestServicePort int            `yaml:"testServiceHostPort" required:"true"`
}

type EventSource struct {
	Postgres *PostgresConfig `yaml:"postgres"`
}

type TemporalConfig struct {
	HostPort  string `yaml:"hostPort" required:"true"`
	Namespace string `yaml:"namespace" required:"true"`
}

type PostgresConfig struct {
	SchemaVersion uint   `yaml:"schemaVersion" required:"true"`
	Host          string `required:"true"`
	Port          string `required:"true"`
	Database      string `required:"true"`
	User          string `required:"true"`
	Password      string `required:"true"`
}

func (p PostgresConfig) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", p.User, p.Password, p.Host, p.Port, p.Database)
}

type DevelopmentConfig struct {
	InMemory bool `yaml:"inMemory"`
	Temporal bool `yaml:"temporal"`
	Logger   bool `yaml:"logger"`
}

type Config interface {
	*AllInOneConfig | *TestServiceConfig | *EventServiceConfig | *WorkflowProxyServiceConfig
}

func LoadConfig[C Config](dst C) error {
	loaderCfg := aconfig.Config{
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
		FileFlag:           "config-file",
		FailOnFileNotFound: true,
	}
	loader := aconfig.LoaderFor(dst, loaderCfg)
	return loader.Load()
}
