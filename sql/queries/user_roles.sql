-- name: CreateUserRoles :one
INSERT INTO user_roles (user_id, role)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserRolesByUserID :many
SELECT * FROM user_roles
WHERE user_id = $1
ORDER BY role ASC;
