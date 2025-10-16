package api

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

func (config *apiConfig) HandlerGetStudents(c echo.Context) error {
	ctx := c.Request().Context()
	q := config.Server.Queries

	students, err := q.GetStudentAll(ctx)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	// do validation caching
	lastModified, err := q.GetCollectionMetaLastModified(ctx, "student-coll")
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

func (config *apiConfig) HandlerGetStudentByID(c echo.Context) error {
	ctx := c.Request().Context()
	qtx := config.Server.Queries
	var param struct {
		ID string `param:"id"`
	}

	if err := c.Bind(&param); err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	if err := c.Validate(&param); err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	id, err := uuid.Parse(param.ID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	student, err := qtx.GetStudentById(ctx, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, studentJSONFormat(student))
}

func (config *apiConfig) HandlerCreateStudent(c echo.Context) error {
	ctx := c.Request().Context()
	qtx := config.Server.Queries

	var reqBody struct {
		Name        string `json:"name" validate:"nochars,cheeky_sql_inject"`
		Email       string `json:"email" validate:"email_constraints,cheeky_sql_inject"`
		PhoneNumber string `json:"phone_number" validate:"phone_constraints"`
		Major       string `json:"major" validate:"oneof_major,nochars"`
		Nip         string `json:"nip" validate:"nip_constraints"`
		DateOfBirth string `json:"date_of_birth" validate:"dob_constraints,cheeky_sql_inject"`
	}

	err := handler.WithTX(ctx, config.Server.DB, qtx, func(qtx *database.Queries) error {
		if err := c.Bind(&reqBody); err != nil {
			return fmt.Errorf("here daddy 63, %v", err.Error())
		}

		if err := c.Validate(&reqBody); err != nil {
			return fmt.Errorf("here daddy 67, %v", err.Error())
		}

		studentBirthDate, err := time.Parse(handler.DOBLayout, reqBody.DateOfBirth)
		if err != nil {
			return fmt.Errorf("here daddy 74, %v", err)
		}

		birthDateStr := fmt.Sprintf("%v", studentBirthDate.Format(time.DateOnly))

		if !handler.IsNIPValid(reqBody.Nip, birthDateStr) {
			return errors.New(handler.ERROR_INVALID_NIP)
		}

		// if free nim exists, get the smallest nim for the new created student
		// and delete the record points to that free nim
		nim, err := qtx.GetFreelistNim(ctx)
		// checks if there is no free nim to be used
		// simply generate from the student-nim
		if err != nil {
			nim, _ = qtx.GetCollectionMetaValue(ctx, "student-nim")
			qtx.IncrementValueByname(ctx, "student-nim")
			log.Println("\n\nstudent creation process...")
		}

		err = qtx.DeleteFreelistNim(ctx, nim)
		if err != nil {
			return fmt.Errorf("here daddy 87, %v", err.Error())
		}

		studentData := c.Get("studentInfo").(*studentData)
		student, err := qtx.CreateStudent(ctx, database.CreateStudentParams{
			Name:        strings.ToLower(reqBody.Name),
			Email:       reqBody.Email,
			PhoneNumber: reqBody.PhoneNumber,
			Nip:         reqBody.Nip,
			Year:        int32(time.Now().Year()),
			Nim:         nim,
			DateOfBirth: studentBirthDate,
			StudyPlanID: studentData.StudyPlan.ID,
			RoomID:      studentData.Room.ID,
		})
		if err != nil {
			return fmt.Errorf("here daddy 102, %v", err.Error())
		}

		// add the student to the classroom
		err = qtx.SetStudentClassroom(ctx, database.SetStudentClassroomParams{
			StudentID: student.ID,
			RoomID:    studentData.Room.ID,
		})
		if err != nil {
			return fmt.Errorf("here daddy 111, %v", err.Error())
		}

		studentClassCount := studentData.StudyPlan.Major + "-StudentCount"
		qtx.IncrementValueByname(ctx, studentClassCount)

		// update lastModifed for validation caching
		if err := qtx.UpdateCollectionMetaLastModified(ctx, "student-coll"); err != nil {
			return fmt.Errorf("here daddy 119, %v", err.Error())
		}

		return nil
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, Data{"status": "succeed, Student Created"})
}

func (config *apiConfig) HandlerDeleteStudent(c echo.Context) error {
	ctx := c.Request().Context()
	q := config.Server.Queries
	var param struct {
		ID uuid.UUID `param:"id"`
	}

	err := handler.WithTX(ctx, config.Server.DB, q, func(qtx *database.Queries) error {
		if err := c.Bind(&param); err != nil {
			return fmt.Errorf("here daddy 142, %v", err.Error())
		}
		if err := c.Validate(&param); err != nil {
			return fmt.Errorf("here daddy 145, %v", err.Error())
		}

		// delete opp
		student, err := qtx.DeleteStudentById(ctx, param.ID)
		if err != nil {
			return fmt.Errorf("here daddy 151, %v", err.Error())
		}

		// decrement the count of student from their class
		studyPlan, err := qtx.GetStudyPlanById(ctx, student.StudyPlanID)
		if err != nil {
			return fmt.Errorf("here daddy 157, %v", err.Error())
		}

		studentClassCount := studyPlan.Major + "-StudentCount"
		if err := qtx.DecrementValueByName(ctx, studentClassCount); err != nil {
			return fmt.Errorf("here daddy 162, %v", err.Error())
		}

		// add their nim to the available nim
		err = qtx.AddToFreelist(ctx, student.Nim)
		if err != nil {
			return fmt.Errorf("here daddy 168, %v", err.Error())
		}

		// update lastModified for caching
		if err := qtx.UpdateCollectionMetaLastModified(ctx, "student-coll"); err != nil {
			return fmt.Errorf("here daddy 173, %v", err.Error())
		}

		return nil
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, Data{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, Data{"succeed": "student get deleted"})
}
