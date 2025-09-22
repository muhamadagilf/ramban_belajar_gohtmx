

-- +goose Up
ALTER TABLE students
ADD COLUMN phone_number VARCHAR(20) NOT NULL CHECK (phone_number ~ '^\+?[0-9]{8,15}$');

-- +goose Down
ALTER TABLE students DROP phone_number;
