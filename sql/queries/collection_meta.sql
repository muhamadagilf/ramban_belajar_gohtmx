
-- name: GetCollectionMetaValue :one
SELECT value FROM collection_meta
WHERE name = $1;

-- name: GetCollectionMetaLastModified :one
SELECT updated_at FROM collection_meta
WHERE name = $1;

-- name: UpdateCollectionMetaLastModified :exec
UPDATE collection_meta
SET updated_at = NOW()
WHERE name = $1;

-- name: GenerateStudentNim :one
UPDATE collection_meta
SET value = (CAST(value as INTEGER)+1)::VARCHAR
WHERE name = 'NIM'
RETURNING *;
