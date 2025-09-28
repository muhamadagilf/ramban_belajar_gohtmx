
-- +goose Up
CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    nip INT UNIQUE NOT NULL,
    name VARCHAR(64) NOT NULL,
    email VARCHAR(64) UNIQUE NOT NULL,
    year INT NOT NULL,
    room_id UUID REFERENCES rooms(id),
    study_plan_id UUID REFERENCES study_plans(id)
);


-- +goose Down
DROP TABLE students;
