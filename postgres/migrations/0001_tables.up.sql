CREATE TABLE contexts
(
    id TEXT PRIMARY KEY
);

CREATE TABLE test_suites
(
    id          UUID NOT NULL PRIMARY KEY,
    context_id  TEXT NOT NULL REFERENCES contexts (id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    UNIQUE (context_id, name)
);

CREATE TABLE test_suite_registrations
(
    context_id    TEXT NOT NULL REFERENCES contexts (id) ON DELETE CASCADE,
    test_suite_id UUID NOT NULL REFERENCES test_suites (id) ON DELETE CASCADE,
    runner_id     TEXT NOT NULL,
    version       TEXT NOT NULL,
    PRIMARY KEY (context_id, test_suite_id)
);

CREATE TABLE tests
(
    id            UUID      NOT NULL PRIMARY KEY,
    context_id    TEXT      NOT NULL REFERENCES contexts (id) ON DELETE CASCADE,
    test_suite_id UUID      NOT NULL REFERENCES test_suites (id) ON DELETE CASCADE,
    name          TEXT      NOT NULL,
    has_input     BOOLEAN   NOT NULL,
    create_time   TIMESTAMP NOT NULL,
    UNIQUE (context_id, test_suite_id, name)
);

CREATE TABLE test_default_inputs
(
    test_id UUID  NOT NULL REFERENCES tests (id) DEFERRABLE PRIMARY KEY,
    data    BYTEA NOT NULL
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