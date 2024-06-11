-- name: CreateGroup :exec
INSERT INTO groups (context_id, name)
VALUES ($1, $2)
ON CONFLICT (context_id, name) DO NOTHING;

-- name: GroupExists :exec
SELECT *
FROM groups
WHERE context_id = $1
  AND name = $2;
