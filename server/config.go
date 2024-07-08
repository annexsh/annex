package server

import (
	"fmt"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type AllInOneConfig struct {
	Port        int               `yaml:"port" required:"true"`
	Postgres    PostgresConfig    `yaml:"postgres"`
	Nats        NatsConfig        `yaml:"nats"`
	Temporal    TemporalConfig    `yaml:"temporal"`
	Development DevelopmentConfig `yaml:"development"`
}

type TestServiceConfig struct {
	Port               int            `yaml:"port" required:"true"`
	Postgres           PostgresConfig `yaml:"postgres"`
	NatsConfig         NatsConfig     `yaml:"nats" required:"true"`
	WorkflowServiceURL string         `yaml:"workflowServiceURL" required:"true"`
}

type EventServiceConfig struct {
	Port           int        `yaml:"port" required:"true"`
	Nats           NatsConfig `yaml:"nats"`
	TestServiceURL string     `yaml:"testServiceURL" required:"true"`
}

type WorkflowProxyServiceConfig struct {
	Port           int            `yaml:"port" required:"true"`
	Temporal       TemporalConfig `yaml:"temporal"`
	TestServiceURL string         `yaml:"testServiceURL" required:"true"`
}

type PostgresConfig struct {
	SchemaVersion uint   `yaml:"schemaVersion" required:"true"`
	HostPort      string `yaml:"hostPort" required:"true"`
	Database      string `yaml:"database" required:"true"`
	User          string `yaml:"user" required:"true"`
	Password      string `yaml:"password" required:"true"`
}

func (p PostgresConfig) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s", p.User, p.Password, p.HostPort, p.Database)
}

type NatsConfig struct {
	HostPort string `yaml:"hostPort" required:"true"`
	// EmbeddedNats embeds a NATs server into the running service. This option
	// only applies to EventServiceConfig and AllInOneConfig.
	EmbeddedNats bool `yaml:"embeddedNats"`
}

type TemporalConfig struct {
	HostPort  string `yaml:"hostPort" required:"true"`
	Namespace string `yaml:"namespace" required:"true"`
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
