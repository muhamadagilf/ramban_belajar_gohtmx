
-- +goose Up
CREATE TABLE students (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    nip INT UNIQUE NOT NULL,
    name VARCHAR(64) NOT NULL,
    email VARCHAR(64) UNIQUE NOT NULL,
    year INT NOT NULL,
    room_id UUID REFERENCES rooms(id),
    study_plan_id UUID REFERENCES study_plans(id)
);


-- +goose Down
DROP TABLE students;
