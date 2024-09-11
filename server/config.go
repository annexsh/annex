package server

import (
	"fmt"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/go-ozzo/ozzo-validation/v4"
)

type AllInOneConfig struct {
	Port        int               `yaml:"port"`
	Postgres    PostgresConfig    `yaml:"postgres"`
	Nats        NatsConfig        `yaml:"nats"`
	Temporal    TemporalConfig    `yaml:"temporal"`
	Development DevelopmentConfig `yaml:"development"`
}

func (c AllInOneConfig) Validate() error {
	if err := validation.ValidateStruct(&c,
		validation.Field(&c.Port, validation.Required, validation.Min(0)),
	); err != nil {
		return err
	}

	if c.Postgres.Empty() && !c.Development.SQLiteDatabase {
		return fmt.Errorf("Postgres required")
	}
	if !c.Postgres.Empty() && c.Development.SQLiteDatabase {
		return fmt.Errorf("Postgres and SQLite are mutually exclusive")
	}
	if !c.Development.SQLiteDatabase {
		if err := c.Postgres.Validate(); err != nil {
			return err
		}
	}
	if err := c.Nats.Validate(); err != nil {
		return err
	}
	if err := c.Temporal.Validate(); err != nil {
		return err
	}

	return nil
}

func LoadAllInOneConfig() (AllInOneConfig, error) {
	cfg := AllInOneConfig{}
	if err := loadConfig(&cfg); err != nil {
		return AllInOneConfig{}, err
	}
	return cfg, nil
}

type TestServiceConfig struct {
	Port               int            `yaml:"port"`
	WorkflowServiceURL string         `yaml:"workflowServiceURL"`
	Postgres           PostgresConfig `yaml:"postgres"`
	Nats               NatsConfig     `yaml:"nats"`
}

func (c TestServiceConfig) Validate() error {
	if err := validation.ValidateStruct(&c,
		validation.Field(&c.Port, validation.Required, validation.Min(0)),
		validation.Field(&c.WorkflowServiceURL, validation.Required),
	); err != nil {
		return err
	}

	if err := c.Postgres.Validate(); err != nil {
		return err
	}
	if err := c.Nats.Validate(); err != nil {
		return err
	}

	return nil
}

func LoadTestServiceConfig() (TestServiceConfig, error) {
	cfg := TestServiceConfig{}
	if err := loadConfig(&cfg); err != nil {
		return TestServiceConfig{}, err
	}
	return cfg, nil
}

type EventServiceConfig struct {
	Port           int        `yaml:"port"`
	TestServiceURL string     `yaml:"testServiceURL"`
	Nats           NatsConfig `yaml:"nats"`
}

func (c EventServiceConfig) Validate() error {
	if err := validation.ValidateStruct(&c,
		validation.Field(&c.Port, validation.Required, validation.Min(0)),
		validation.Field(&c.TestServiceURL, validation.Required),
	); err != nil {
		return err
	}

	if err := c.Nats.Validate(); err != nil {
		return err
	}

	return nil
}

func LoadEventServiceConfig() (EventServiceConfig, error) {
	cfg := EventServiceConfig{}
	if err := loadConfig(&cfg); err != nil {
		return EventServiceConfig{}, err
	}
	return cfg, nil
}

type WorkflowProxyServiceConfig struct {
	Port           int            `yaml:"port"`
	TestServiceURL string         `yaml:"testServiceURL"`
	Temporal       TemporalConfig `yaml:"temporal"`
}

func (c WorkflowProxyServiceConfig) Validate() error {
	if err := validation.ValidateStruct(&c,
		validation.Field(&c.Port, validation.Required, validation.Min(0)),
		validation.Field(&c.TestServiceURL, validation.Required),
	); err != nil {
		return err
	}

	if err := c.Temporal.Validate(); err != nil {
		return err
	}

	return nil
}

func LoadWorkflowProxyServiceConfig() (WorkflowProxyServiceConfig, error) {
	cfg := WorkflowProxyServiceConfig{}
	if err := loadConfig(&cfg); err != nil {
		return WorkflowProxyServiceConfig{}, err
	}
	return cfg, nil
}

type PostgresConfig struct {
	HostPort string `yaml:"hostPort"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func (c PostgresConfig) Validate() error {
	if err := validation.ValidateStruct(&c,
		validation.Field(&c.HostPort, validation.Required),
		validation.Field(&c.User, validation.Required),
		validation.Field(&c.Password, validation.Required),
	); err != nil {
		return fmt.Errorf("Postgres: %w", err)
	}
	return nil
}

func (c PostgresConfig) Empty() bool {
	return c == PostgresConfig{}
}

type NatsConfig struct {
	HostPort string `yaml:"hostPort"`
	// Embedded embeds a NATs server into the running service.
	// This option only applies to EventServiceConfig and AllInOneConfig.
	Embedded bool `yaml:"embedded"`
}

func (c NatsConfig) Validate() error {
	if err := validation.ValidateStruct(&c,
		validation.Field(&c.HostPort, validation.Required),
	); err != nil {
		return fmt.Errorf("Nats: %w", err)
	}
	return nil
}

type TemporalConfig struct {
	HostPort  string `yaml:"hostPort"`
	Namespace string `yaml:"namespace"`
}

func (c TemporalConfig) Validate() error {
	if err := validation.ValidateStruct(&c,
		validation.Field(&c.HostPort, validation.Required),
		validation.Field(&c.Namespace, validation.Required),
	); err != nil {
		return fmt.Errorf("Temporal: %w", err)
	}
	return nil
}

type DevelopmentConfig struct {
	Logger           bool `yaml:"logger"`
	EmbeddedTemporal bool `yaml:"embeddedTemporal"`
	SQLiteDatabase   bool `yaml:"sqliteDatabase"`
}

type validator interface {
	Validate() error
}

func loadConfig(dst validator) error {
	loaderCfg := aconfig.Config{
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
		FileFlag:           "config-file",
		FailOnFileNotFound: true,
	}

	loader := aconfig.LoaderFor(dst, loaderCfg)
	if err := loader.Load(); err != nil {
		return err
	}

	if err := dst.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	return nil
}
