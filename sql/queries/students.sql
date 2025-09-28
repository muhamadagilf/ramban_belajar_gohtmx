
-- name: CreateStudent :one
INSERT INTO students (nip, name, email, year, room_id, study_plan_id, phone_number, nim)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetStudentAll :many
SELECT * FROM students
ORDER BY updated_at DESC;

-- name: GetStudentById :one
SELECT * FROM students
WHERE id = $1;

-- name: GetRecentCreatedStudent :one
SELECT * FROM students
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteStudentById :exec
DELETE FROM students
WHERE id = $1;

-- name: UpdateStudent :one
UPDATE students
SET email = $2, phone_number = $3, updated_at = $4
WHERE id = $1
RETURNING *;
