CREATE TABLE contexts
(
    id TEXT PRIMARY KEY
);

CREATE TABLE test_suites
(
    id          TEXT NOT NULL PRIMARY KEY,
    context_id  TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT,
    UNIQUE (context_id, name),
    FOREIGN KEY (context_id) REFERENCES contexts (id) ON DELETE CASCADE
);

CREATE TABLE test_suite_registrations
(
    context_id    TEXT NOT NULL,
    test_suite_id TEXT NOT NULL,
    runner_id     TEXT NOT NULL,
    version       TEXT NOT NULL,
    PRIMARY KEY (context_id, test_suite_id),
    FOREIGN KEY (context_id) REFERENCES contexts (id) ON DELETE CASCADE,
    FOREIGN KEY (test_suite_id) REFERENCES test_suites (id) ON DELETE CASCADE
);

CREATE TABLE tests
(
    id            TEXT     NOT NULL PRIMARY KEY,
    context_id    TEXT     NOT NULL,
    test_suite_id TEXT     NOT NULL,
    name          TEXT     NOT NULL,
    has_input     BOOLEAN  NOT NULL,
    create_time   DATETIME NOT NULL,
    FOREIGN KEY (context_id) REFERENCES contexts (id) ON DELETE CASCADE,
    FOREIGN KEY (test_suite_id) REFERENCES test_suites (id) ON DELETE CASCADE,
    UNIQUE (context_id, test_suite_id, name)
);

CREATE TABLE test_default_inputs
(
    test_id TEXT NOT NULL PRIMARY KEY,
    data    BLOB NOT NULL,
    FOREIGN KEY (test_id) REFERENCES tests (id) DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE test_executions
(
    id            TEXT     NOT NULL PRIMARY KEY,
    test_id       TEXT     NOT NULL,
    has_input     BOOLEAN  NOT NULL,
    schedule_time DATETIME NOT NULL,
    start_time    DATETIME,
    finish_time   DATETIME,
    error         TEXT,
    FOREIGN KEY (test_id) REFERENCES tests (id)
);

CREATE TABLE test_execution_inputs
(
    test_execution_id TEXT NOT NULL PRIMARY KEY,
    data              BLOB NOT NULL,
    FOREIGN KEY (test_execution_id) REFERENCES test_executions (id)
);

CREATE TABLE case_executions
(
    id                INTEGER  NOT NULL,
    test_execution_id TEXT     NOT NULL,
    case_name         TEXT     NOT NULL,
    schedule_time     DATETIME NOT NULL,
    start_time        DATETIME,
    finish_time       DATETIME,
    error             TEXT,
    PRIMARY KEY (id, test_execution_id),
    FOREIGN KEY (test_execution_id) REFERENCES test_executions (id)
);

CREATE TABLE logs
(
    id                TEXT PRIMARY KEY,
    test_execution_id TEXT     NOT NULL,
    case_execution_id INTEGER,
    level             TEXT     NOT NULL,
    message           TEXT     NOT NULL,
    create_time       DATETIME NOT NULL,
    FOREIGN KEY (test_execution_id) REFERENCES test_executions (id)
);
