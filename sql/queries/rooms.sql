

-- name: GetStudentRoom :many
SELECT * FROM rooms
WHERE name LIKE $1;

-- name: GetStudentRoomById :one
SELECT * FROM rooms
WHERE id = $1;
