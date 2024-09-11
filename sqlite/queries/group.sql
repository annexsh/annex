-- name: CreateGroup :exec
INSERT INTO groups (context_id, id)
VALUES (?, ?)
ON CONFLICT (context_id, id) DO NOTHING;

-- name: ListGroups :many
SELECT *
FROM groups
WHERE (context_id = @context_id)
  AND (id > COALESCE(sqlc.narg('offset_id'), ''))
ORDER BY id
LIMIT @page_size;


-- name: GroupExists :exec
SELECT *
FROM groups
WHERE context_id = ?
  AND id = ?;
