-- name: CreateContext :exec
INSERT INTO contexts (id)
VALUES ($1);


-- name: ListContexts :many
SELECT *
FROM contexts
WHERE (id > COALESCE(sqlc.narg('offset_id'), ''))
ORDER BY id
LIMIT @page_size;
