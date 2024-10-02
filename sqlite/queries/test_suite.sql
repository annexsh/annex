-- name: CreateTestSuite :one
INSERT INTO test_suites (id, context_id, name, description)
VALUES (?, ?, ?, ?)
ON CONFLICT (context_id, name) DO UPDATE
    SET description = excluded.description
RETURNING id;;

-- name: ListTestSuites :many
SELECT *
FROM test_suites
WHERE (context_id = @context_id)
  AND (name > COALESCE(sqlc.narg('offset_id'), ''))
ORDER BY name
LIMIT @page_size;

-- name: SetTestSuiteVersion :exec
INSERT INTO test_suite_registrations (context_id, test_suite_id, runner_id, version)
VALUES (?, ?, ?, ?)
ON CONFLICT (context_id, test_suite_id) DO UPDATE
    SET runner_id = excluded.runner_id,
        version   = excluded.version;

-- name: GetTestSuiteVersion :one
SELECT version
FROM test_suite_registrations
WHERE context_id = ?
  AND test_suite_id = ?;
