-- name: CreateTest :one
INSERT INTO tests (context_id, group_id, id, name, has_input)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (context_id, group_id, name) DO UPDATE
    SET has_input   = excluded.has_input,
        create_time = now()
RETURNING *;

-- name: GetTest :one
SELECT *
FROM tests
WHERE id = $1;

-- name: GetTestByName :one
SELECT *
FROM tests
WHERE name = $1
  AND group_id = $2;

-- name: ListTests :many
SELECT *
FROM tests
WHERE context_id = $1 AND group_id = $2;

-- name: CreateTestDefaultInput :exec
INSERT INTO test_default_inputs (test_id, data)
VALUES ($1, $2)
ON CONFLICT (test_id) DO UPDATE
    SET data = excluded.data;

-- name: GetTestDefaultInput :one
SELECT *
FROM test_default_inputs
WHERE test_id = $1;
