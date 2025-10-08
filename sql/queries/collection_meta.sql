
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

-- name: IncrementValueByname :exec
UPDATE collection_meta
SET value = (CAST(value as INTEGER)+1)::VARCHAR
WHERE name = $1;

-- name: DecrementValueByName :exec
UPDATE collection_meta
SET value = (CAST(value as INTEGER)-1)::VARCHAR
WHERE name = $1;

-- name: GetFreelistNim :one
SELECT value FROM collection_meta
WHERE name = 'freelist-nim'
ORDER BY value ASC
LIMIT 1;

-- name: DeleteFreelistNim :exec
DELETE FROM collection_meta
WHERE name = 'freelist-nim' AND value = $1;

-- name: AddToFreelist :exec
INSERT INTO collection_meta (name, value)
VALUES ('freelist-nim', $1);