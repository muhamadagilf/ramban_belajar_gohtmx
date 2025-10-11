
-- +goose Up
ALTER TABLE students
ALTER COLUMN nip TYPE VARCHAR(24);

-- +goose Down
ALTER TABLE students
ALTER COLUMN nip TYPE INT;
