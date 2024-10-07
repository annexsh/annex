-- name: CreateTest :one
INSERT INTO tests (context_id, test_suite_id, id, name, has_input, create_time)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(context_id, test_suite_id, name) DO UPDATE
    SET has_input   = excluded.has_input
RETURNING *;

-- name: GetTest :one
SELECT *
FROM tests
WHERE id = ?;

-- name: GetTestByName :one
SELECT *
FROM tests
WHERE name = ?
  AND test_suite_id = ?;

-- name: ListTests :many
SELECT *
FROM tests
WHERE (context_id = @context_id AND test_suite_id = @test_suite_id)
  -- Cast as text required below since sqlc.narg doesn't work with overridden column type
  AND (CAST(sqlc.narg('offset_id') AS TEXT) IS NULL OR id < CAST(sqlc.narg('offset_id') AS TEXT))
ORDER BY id DESC
LIMIT @page_size;

-- name: DeleteTest :exec
DELETE
FROM tests
WHERE id = ?;

-- name: CreateTestDefaultInput :exec
INSERT INTO test_default_inputs (test_id, data)
VALUES (?, ?)
ON CONFLICT(test_id) DO UPDATE
    SET data = excluded.data;

-- name: GetTestDefaultInput :one
SELECT *
FROM test_default_inputs
WHERE test_id = ?;
