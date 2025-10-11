package api

import (
	"time"

	"github.com/google/uuid"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

type StudentFormat struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Nip         string    `json:"nip"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Year        int32     `json:"year"`
	RoomID      uuid.UUID `json:"room_id"`
	StudyPlanID uuid.UUID `json:"study_plan_id"`
	PhoneNumber string    `json:"phone_number"`
	Nim         string    `json:"nim"`
	DateOfBirth string    `json:"date_of_birth"`
}

func studentJSONFormat(student database.Student) StudentFormat {
	return StudentFormat{
		student.ID,
		student.CreatedAt,
		student.UpdatedAt,
		student.Nip,
		student.Name,
		student.Email,
		student.Year,
		student.RoomID,
		student.StudyPlanID,
		student.PhoneNumber,
		student.Nim,
		student.DateOfBirth.Format(time.DateOnly),
	}
}

func studentsJSONFormat(students []database.Student) []StudentFormat {
	s := []StudentFormat{}
	for _, v := range students {
		s = append(s, studentJSONFormat(v))
	}

	return s
}
