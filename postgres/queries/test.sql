-- name: CreateTest :one
INSERT INTO tests (context_id, test_suite_id, id, name, has_input, create_time)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (context_id, test_suite_id, name) DO UPDATE
    SET has_input = excluded.has_input
RETURNING *;

-- name: GetTest :one
SELECT *
FROM tests
WHERE id = $1;

-- name: GetTestByName :one
SELECT *
FROM tests
WHERE name = $1
  AND test_suite_id = $2;

-- name: ListTests :many
SELECT *
FROM tests
WHERE (context_id = @context_id AND test_suite_id = @test_suite_id)
  AND (sqlc.narg('offset_id')::uuid IS NULL OR id < sqlc.narg('offset_id')::uuid)
ORDER BY id DESC
LIMIT @page_size;

-- name: DeleteTest :exec
DELETE
FROM tests
WHERE id = $1;

-- name: CreateTestDefaultInput :exec
INSERT INTO test_default_inputs (test_id, data)
VALUES ($1, $2)
ON CONFLICT (test_id) DO UPDATE
    SET data = excluded.data;

-- name: GetTestDefaultInput :one
SELECT *
FROM test_default_inputs
WHERE test_id = $1;
