-- name: CreateTestSuite :one
INSERT INTO test_suites (id, context_id, name, description)
VALUES ($1, $2, $3, $4)
ON CONFLICT (context_id, name) DO UPDATE
    SET description = excluded.description
RETURNING id;

-- name: ListTestSuites :many
SELECT sqlc.embed(ts), COUNT(t.id) AS test_count
FROM test_suites ts
         LEFT JOIN
     tests t ON ts.id = t.test_suite_id
WHERE (ts.context_id = @context_id)
  AND (ts.name > COALESCE(sqlc.narg('offset_id'), ''))
GROUP BY ts.id
ORDER BY ts.name
LIMIT @page_size;

SELECT ts.*, COUNT(t.id) AS test_count
FROM test_suites ts
         LEFT JOIN
     tests t ON ts.id = t.test_suite_id
GROUP BY ts.id;

-- name: SetTestSuiteVersion :exec
INSERT INTO test_suite_registrations (context_id, test_suite_id, runner_id, version)
VALUES ($1, $2, $3, $4)
ON CONFLICT (context_id, test_suite_id) DO UPDATE
    SET runner_id = excluded.runner_id,
        version   = excluded.version;

-- name: GetTestSuiteVersion :one
SELECT version
FROM test_suite_registrations
WHERE context_id = $1
  AND test_suite_id = $2 FOR UPDATE;
