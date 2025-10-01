
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

-- name: DeleteStudentById :one
DELETE FROM students
WHERE id = $1
RETURNING *;

-- name: UpdateStudent :one
UPDATE students
SET email = $2, phone_number = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: GetStudentsByRoomAndMajor :many
SELECT s.id, s.created_at, s.updated_at, s.name, s.email, s.nim, 
s.phone_number, r.name as room, std.major 
FROM students as s
JOIN rooms as r
        ON s.room_id = r.id
JOIN study_plans as std
	ON s.study_plan_id = std.id
WHERE std.major = $1 OR r.name = $2;