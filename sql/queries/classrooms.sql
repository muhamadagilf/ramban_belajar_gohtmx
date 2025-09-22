

-- name: SetStudentClassroom :exec
INSERT INTO classrooms (room_id, student_id)
VALUES ($1, $2);
