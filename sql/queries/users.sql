-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUsersAll :many
SELECT * FROM users;

-- name: GetUsersAllJoinRoles :many
SELECT u.id, u.email, r.role, u.created_at
FROM users AS u
JOIN user_roles as r
  ON u.id = r.user_id;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = $1;
