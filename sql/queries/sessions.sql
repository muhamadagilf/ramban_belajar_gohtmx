-- name: CreateUserSession :one
INSERT INTO sessions (session_id, user_id, expire_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteUserSession :exec
DELETE FROM sessions
WHERE session_id = $1;

-- name: GetUserSession :one
SELECT * FROM sessions
WHERE session_id = $1;

-- name: UpdateLastActivityUserSession :exec
UPDATE sessions
SET last_activity = NOW()
WHERE session_id = $1;

-- name: UpdateRevokeStatusUserSession :exec
UPDATE sessions
SET is_revoked = $1
WHERE session_id = $1;

-- name: CleanupRevokedSessions :exec
DELETE FROM sessions
WHERE is_revoked = true OR expire_at < NOW();
