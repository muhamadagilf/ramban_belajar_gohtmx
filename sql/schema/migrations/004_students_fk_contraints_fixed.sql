

-- +goose Up
ALTER TABLE students
    ALTER COLUMN room_id SET NOT NULL,
    ALTER COLUMN study_plan_id SET NOT NULL;


-- +goose Down
ALTER TABLE students
    ALTER COLUMN room_id DROP NOT NULL,
    ALTER COLUMN study_plan_id DROP NOT NULL;
