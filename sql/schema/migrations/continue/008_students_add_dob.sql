-- +goose Up
ALTER TABLE students
ADD COLUMN date_of_birth DATE NOT NULL;

-- +goose Down
ALTER TABLE students DROP date_of_birth;
