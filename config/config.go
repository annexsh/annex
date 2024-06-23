package config

import (
	"fmt"
)

type Type string

const (
	TypeAll           = "all"
	TypeTest          = "test"
	TypeEvent         = "event"
	TypeWorkflowProxy = "workflow-proxy"
)

type AllServices struct {
	Port        int         `yaml:"port" required:"true"`
	Repository  Repository  `yaml:"repository"`
	EventSource EventSource `yaml:"eventSource"`
	Temporal    Temporal    `yaml:"temporal"`
	Development Development `yaml:"development"`
}

type TestService struct {
	Port                int        `yaml:"port" required:"true"`
	Repository          Repository `yaml:"repository"`
	WorkflowServicePort int        `yaml:"workflowHostPort" required:"true"`
}

type EventService struct {
	Port            int         `yaml:"port" required:"true"`
	EventSource     EventSource `yaml:"eventSource"`
	TestServicePort string      `yaml:"testServiceHostPort" required:"true"`
}

type WorkflowProxyService struct {
	Port            int      `yaml:"port" required:"true"`
	Temporal        Temporal `yaml:"temporal"`
	TestServicePort int      `yaml:"testServiceHostPort" required:"true"`
}

type Repository struct {
	Postgres *Postgres `yaml:"postgres"`
}

type EventSource struct {
	Postgres *Postgres `yaml:"postgres"`
}

type Temporal struct {
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

type Development struct {
	InMemory bool `yaml:"inMemory"`
	Temporal bool `yaml:"temporal"`
	Logger   bool `yaml:"logger"`
}
