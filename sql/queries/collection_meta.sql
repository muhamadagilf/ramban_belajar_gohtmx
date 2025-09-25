
-- name: GetCollectionMeta :one
SELECT updated_at FROM collection_meta
WHERE name = $1;

-- name: UpdateCollectionMeta :exec
UPDATE collection_meta
SET updated_at = NOW()
WHERE name = $1;
