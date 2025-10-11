-- +goose Up
ALTER TABLE students
ADD COLUMN user_id UUID REFERENCES users(id) NOT NULL;

-- +goose Down
ALTER TABLE students DROP user_id;