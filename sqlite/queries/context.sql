-- name: CreateContext :exec
INSERT INTO contexts (id)
VALUES (?);

-- name: ListContexts :many
SELECT *
FROM contexts
WHERE (id > COALESCE(sqlc.narg('offset_id'), ''))
ORDER BY id
LIMIT @page_size;
