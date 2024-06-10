CREATE TABLE groups
(
    id   UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE tests
(
    group_id            UUID                    NOT NULL REFERENCES groups (id),
    id                  UUID                    NOT NULL,
    name                TEXT                    NOT NULL,
    has_input           BOOLEAN                 NOT NULL,
    runner_id           TEXT                    NOT NULL,
    runner_heartbeat_at TIMESTAMP               NOT NULL,
    create_time         TIMESTAMP DEFAULT now() NOT NULL,
    PRIMARY KEY (group_id, id)
);

CREATE TABLE test_default_inputs
(
    group_id UUID  NOT NULL REFERENCES groups (id),
    test_id  UUID  NOT NULL REFERENCES tests (id) DEFERRABLE,
    data     BYTEA NOT NULL,
    PRIMARY KEY (group_id, test_id)
);

CREATE TABLE runners
(
    group_id UUID  NOT NULL REFERENCES groups (id),
    id TEXT NOT NULL,
    test_id   UUID NOT NULL REFERENCES tests (id) DEFERRABLE,
    PRIMARY KEY (runner_id, test_id)
);

CREATE TABLE test_executions
(
    id            UUID      NOT NULL PRIMARY KEY,
    test_id       UUID      NOT NULL REFERENCES tests (id),
    has_input     BOOLEAN   NOT NULL,
    schedule_time TIMESTAMP NOT NULL,
    start_time    TIMESTAMP,
    finish_time   TIMESTAMP,
    error         TEXT
);

CREATE TABLE test_execution_inputs
(
    test_execution_id UUID  NOT NULL REFERENCES test_executions (id) DEFERRABLE PRIMARY KEY,
    data              BYTEA NOT NULL
);

CREATE TABLE case_executions
(
    id                INTEGER   NOT NULL,
    test_execution_id UUID      NOT NULL REFERENCES test_executions (id),
    case_name         TEXT      NOT NULL,
    schedule_time     TIMESTAMP NOT NULL,
    start_time        TIMESTAMP,
    finish_time       TIMESTAMP,
    error             TEXT,
    PRIMARY KEY (id, test_execution_id)
);

CREATE TABLE logs
(
    id                UUID PRIMARY KEY,
    test_execution_id UUID      NOT NULL REFERENCES test_executions (id),
    case_execution_id INTEGER,
    level             TEXT      NOT NULL,
    message           TEXT      NOT NULL,
    create_time       TIMESTAMP NOT NULL
);