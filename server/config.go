package server

import (
	"fmt"

	"github.com/cohesivestack/valgo"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"

	"github.com/annexsh/annex/internal/validator"
)

type AllInOneConfig struct {
	Port        int               `yaml:"port"`
	Postgres    PostgresConfig    `yaml:"postgres"`
	Nats        NatsConfig        `yaml:"nats"`
	Temporal    TemporalConfig    `yaml:"temporal"`
	Development DevelopmentConfig `yaml:"development"`
}

func (c AllInOneConfig) Validate() error {
	v := validator.New(validator.WithBaseErrorMessage("invalid config"))
	v.Is(valgo.Int(c.Port, "port").GreaterOrEqualTo(0))
	if c.Postgres.Empty() && !c.Development.SQLiteDatabase {
		v.AddErrorMessage("postgres", "Postgres configuration required")
	}
	if !c.Postgres.Empty() && c.Development.SQLiteDatabase {
		v.AddErrorMessage("postgres", "Postgres and SQLite configuration are mutually exclusive")
	}
	if !c.Development.SQLiteDatabase {
		v.In("postgres", c.Postgres.Validation())
	}
	v.In("nats", c.Nats.Validation())
	v.In("temporal", c.Temporal.Validation())
	return v.Error()
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
	v := validator.New(validator.WithBaseErrorMessage("invalid config"))
	v.Is(valgo.Int(c.Port, "port").GreaterOrEqualTo(0))
	v.In("postgres", c.Postgres.Validation())
	v.In("nats", c.Nats.Validation())
	return v.Error()
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
	v := validator.New()
	v.Is(
		valgo.Int(c.Port, "port").GreaterOrEqualTo(0),
		valgo.String(c.TestServiceURL, "testServiceURL").Not().Blank(),
	)
	v.In("nats", c.Nats.Validation())
	return v.Error()
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
	v := validator.New(validator.WithBaseErrorMessage("invalid config"))
	v.Is(
		valgo.Int(c.Port, "port").GreaterOrEqualTo(0),
		valgo.String(c.TestServiceURL, "testServiceURL").Not().Blank(),
	)
	v.In("temporal", c.Temporal.Validation())
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

func (c PostgresConfig) Validation() *valgo.Validation {
	return valgo.Is(
		validator.HostPort(c.HostPort, "hostPort"),
		valgo.String(c.User, "user").Not().Blank(),
	)
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

func (c NatsConfig) Validation() *valgo.Validation {
	return valgo.Is(validator.HostPort(c.HostPort, "hostPort"))
}

type TemporalConfig struct {
	HostPort  string `yaml:"hostPort"`
	Namespace string `yaml:"namespace"`
}

func (c TemporalConfig) Validation() *valgo.Validation {
	return valgo.Is(
		validator.HostPort(c.HostPort, "hostPort"),
		valgo.String(c.Namespace, "namespace").Not().Blank(),
	)
}

type DevelopmentConfig struct {
	Logger           bool `yaml:"logger"`
	EmbeddedTemporal bool `yaml:"embeddedTemporal"`
	SQLiteDatabase   bool `yaml:"sqliteDatabase"`
}

type configValidator interface {
	Validate() error
}

func loadConfig(dst configValidator) error {
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
