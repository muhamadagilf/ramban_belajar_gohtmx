
-- +goose Up
ALTER TABLE students
ADD COLUMN nim VARCHAR(20) UNIQUE NOT NULL;

-- +goose Down
ALTER TABLE students DROP COLUMN nim;
