package api

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

type StudentFormat struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Nip         int32     `json:"nip"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Year        int32     `json:"year"`
	RoomID      uuid.UUID `json:"room_id"`
	StudyPlanID uuid.UUID `json:"study_plan_id"`
	PhoneNumber string    `json:"phone_number"`
	Nim         string    `json:"nim"`
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
	}
}

func studentsJSONFormat(students []database.Student) []StudentFormat {
	s := []StudentFormat{}
	for _, v := range students {
		s = append(s, studentJSONFormat(v))
	}

	return s
}

type Data = map[string]interface{}

func (config *apiConfig) HandlerHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, Data{"Message": "from :3000 up and running..."})
}

func (config *apiConfig) HandlerGetStudents(c echo.Context) error {
	ctx := c.Request().Context()
	qtx := config.Server.Queries

	students, err := qtx.GetStudentAll(ctx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	// do validation caching
	lastModified, err := qtx.GetCollectionMetaLastModified(ctx, "student-coll")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Data{"error": err.Error()})
	}

	ETag := fmt.Sprintf("%x", sha256.Sum256([]byte(lastModified.Format(time.RFC3339))))

	modifiedSince := c.Request().Header.Get("If-Modifed-Since")
	if c.Request().Header.Get("If-None-Match") == ETag || handler.IsLastModifiedValid(modifiedSince, lastModified) {
		return c.NoContent(http.StatusNotModified)
	}

	c.Response().Header().Set("ETag", ETag)
	c.Response().Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	c.Response().Header().Set("Cache-Control", "no-cache")

	return c.JSON(http.StatusOK, studentsJSONFormat(students))
}
