-- name: CreateContext :exec
INSERT INTO contexts (id)
VALUES ($1);

-- name: ContextExists :exec
SELECT *
FROM contexts
WHERE id = $1;