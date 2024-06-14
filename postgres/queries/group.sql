-- name: CreateGroup :exec
INSERT INTO groups (context_id, id)
VALUES ($1, $2)
ON CONFLICT (context_id, id) DO NOTHING;

-- name: ListGroups :many
SELECT id
FROM groups
WHERE context_id = $1;

-- name: GroupExists :exec
SELECT *
FROM groups
WHERE context_id = $1
  AND id = $2;
