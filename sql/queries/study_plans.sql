
-- name: GetStudyPlan :one
SELECT * FROM study_plans
WHERE semester = $1 AND major = $2;

-- name: GetStudyPlanById :one
SELECT * FROM study_plans
WHERE id = $1;
