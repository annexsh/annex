CREATE TABLE tests
(
    id                  UUID PRIMARY KEY,
    project             TEXT                    NOT NULL,
    name                TEXT                    NOT NULL,
    has_payload         BOOLEAN                 NOT NULL,
    runner_id           TEXT                    NOT NULL,
    runner_heartbeat_at TIMESTAMP               NOT NULL,
    created_at          TIMESTAMP DEFAULT now() NOT NULL,
    UNIQUE (project, name)
);

CREATE TABLE test_default_payloads
(
    test_id UUID    NOT NULL REFERENCES tests (id) DEFERRABLE PRIMARY KEY,
    payload BYTEA   NOT NULL,
    is_zero BOOLEAN NOT NULL
);

CREATE TABLE test_executions
(
    id           UUID      NOT NULL PRIMARY KEY,
    test_id      UUID      NOT NULL REFERENCES tests (id),
    has_payload  BOOLEAN   NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    started_at   TIMESTAMP,
    finished_at  TIMESTAMP,
    error        TEXT
);

CREATE TABLE test_execution_payloads
(
    test_exec_id UUID  NOT NULL REFERENCES test_executions (id) DEFERRABLE PRIMARY KEY,
    payload      JSONB NOT NULL
);

CREATE TABLE case_executions
(
    id           INTEGER   NOT NULL,
    test_exec_id UUID      NOT NULL REFERENCES test_executions (id),
    case_name    TEXT      NOT NULL,
    scheduled_at TIMESTAMP NOT NULL,
    started_at   TIMESTAMP,
    finished_at  TIMESTAMP,
    error        TEXT,
    PRIMARY KEY (id, test_exec_id)
);

CREATE TABLE logs
(
    id           UUID PRIMARY KEY,
    test_exec_id UUID      NOT NULL REFERENCES test_executions (id),
    case_exec_id INTEGER,
    level        TEXT      NOT NULL,
    message      TEXT      NOT NULL,
    created_at   TIMESTAMP NOT NULL
);