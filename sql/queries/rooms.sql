

-- name: GetStudentRoom :many
SELECT * FROM rooms
WHERE name LIKE $1;
