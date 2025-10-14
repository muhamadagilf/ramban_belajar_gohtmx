// Package web HandlerFunc
package web

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

type Data map[string]any

func (config *webConfig) GetStudentSubmitPage(c echo.Context) error {
	return c.Render(http.StatusOK, "student-submission", Data{
		"Major": MAJOR,
	})
}

func (config *webConfig) CreateStudent(c echo.Context) error {
	time.Sleep(300 * time.Millisecond)
	ctx := c.Request().Context()
	type formParams struct {
		Name            string `validate:"name_constraints,cheeky_sql_inject"`
		Email           string `validate:"email_constraints,cheeky_sql_inject"`
		PhoneNumber     string `validate:"phone_constraints"`
		Nip             string `validate:"nip_constraints"`
		DateOfBirth     string `validate:"cheeky_sql_inject"`
		Password        string
		ConfirmPassword string
	}

	err := handler.WithTX(ctx, config.Server.DB, config.Server.Queries, func(qtx *database.Queries) error {
		params := formParams{
			Name:            c.FormValue("fullname"),
			Email:           c.FormValue("email"),
			PhoneNumber:     c.FormValue("phone"),
			Nip:             c.FormValue("nip"),
			DateOfBirth:     c.FormValue("birthdate"),
			Password:        c.FormValue("password"),
			ConfirmPassword: c.FormValue("confirm-password"),
		}

		if err := c.Validate(&params); err != nil {
			return err
		}

		if params.Password != params.ConfirmPassword {
			return errors.New(handler.ERROR_INVALID_CONFIRM_PASSWORD)
		}

		if !handler.IsNIPValid(params.Nip, params.DateOfBirth) {
			return errors.New(handler.ERROR_INVALID_NIP)
		}

		studentBirthDate, err := time.Parse(time.DateOnly, params.DateOfBirth)
		if err != nil {
			return err
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
			return err
		}

		// hash the user password
		hashedPassword, err := handler.HashPassword(params.Password)
		if err != nil {
			return err
		}

		// doing user & student creation
		user, err := qtx.CreateUser(ctx, database.CreateUserParams{
			Email:        params.Email,
			PasswordHash: hashedPassword,
			Role:         handler.USER_ROLE_STUDENT,
		})
		if err != nil {
			return err
		}

		studentDat := c.Get("studentData").(*StudentData)
		student, err := qtx.CreateStudent(ctx, database.CreateStudentParams{
			Nim:         nim,
			Nip:         params.Nip,
			Name:        strings.ToLower(params.Name),
			Email:       params.Email,
			PhoneNumber: params.PhoneNumber,
			DateOfBirth: studentBirthDate,
			Year:        int32(YEAR),
			StudyPlanID: studentDat.StudyPlan.ID,
			RoomID:      studentDat.Room.ID,
			UserID:      user.ID,
		})
		if err != nil {
			return err
		}

		// add the student to the classroom
		err = qtx.SetStudentClassroom(ctx, database.SetStudentClassroomParams{
			RoomID:    student.RoomID,
			StudentID: student.ID,
		})
		if err != nil {
			return err
		}

		// updated the collection_meta "-StudentCount", increment the count
		// to keep track of the member of class
		studentClassCount := studentDat.StudyPlan.Major + "-StudentCount"
		qtx.IncrementValueByname(ctx, studentClassCount)

		// update updated_at for Last-Modified Header (caching)
		qtx.UpdateCollectionMetaLastModified(ctx, "student-coll")

		return nil
	})
	if err != nil {
		c.Render(http.StatusUnprocessableEntity, "student-submission", Data{})
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": handler.SubmissionErrorMsg(err.Error()),
		})
	}

	c.Response().Header().Set("HX-Redirect", "/students")
	return c.NoContent(http.StatusCreated)
}

func (config *webConfig) GetStudentsPage(c echo.Context) error {
	ctx := c.Request().Context()

	// retrieves all the necessary data, including query params handling
	// do some filter & search querying
	studentsPageData, err := studentsQueryParamHandler(c, config.Server.Queries)
	if err != nil {
		return c.String(
			http.StatusBadRequest,
			fmt.Sprintf("Dont tryna act smart now, you cheeky bastard. \n%v",
				err.Error()),
		)
	}

	studentsPageData["Rooms"] = ROOM
	studentsPageData["Majors"] = MAJOR

	// do validation based caching
	lastModified, err := config.Server.Queries.GetCollectionMetaLastModified(ctx, "student-coll")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	ETag := fmt.Sprintf("%x", sha256.Sum256([]byte(lastModified.Format(time.RFC3339))))

	modifiedSince := c.Request().Header.Get("If-Modified-Since")
	if c.Request().Header.Get("If-None-Match") == ETag || handler.IsLastModifiedValid(modifiedSince, lastModified) {
		return c.NoContent(http.StatusNotModified)
	}

	c.Response().Header().Set("ETag", ETag)
	c.Response().Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	c.Response().Header().Set("Cache-Control", "no-cache")

	return c.Render(http.StatusOK, "students", studentsPageData)
}

func (config *webConfig) DeleteStudent(c echo.Context) error {
	time.Sleep(300 * time.Millisecond)
	ctx := c.Request().Context()
	err := handler.WithTX(ctx, config.Server.DB, config.Server.Queries, func(qtx *database.Queries) error {
		idStr := c.Param("id")

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		student, err := qtx.DeleteStudentById(ctx, id)
		if err != nil {
			return err
		}

		// updated the collection_meta "-StudentCount"
		// decrement the student count, if there is deletion
		studentPlan, _ := qtx.GetStudyPlanById(ctx, student.StudyPlanID)
		studentClassCount := studentPlan.Major + "-StudentCount"
		qtx.DecrementValueByName(ctx, studentClassCount)

		// simply, add the nim to freelist, if there is student deletion
		qtx.AddToFreelist(ctx, student.Nim)

		// update updated_at for Last-Modified Header (caching)
		err = qtx.UpdateCollectionMetaLastModified(ctx, "student-coll")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		c.Render(http.StatusUnprocessableEntity, "students", Data{})
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": err.Error(),
		})
	}

	c.Response().Header().Set("HX-Redirect", "/students")
	return c.Render(http.StatusSeeOther, "completion-message", Data{"Message": "Deletion Complete"})
}

func (config *webConfig) GetStudentProfile(c echo.Context) error {
	// retrieve all necessary data
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	student, err := config.Server.Queries.GetStudentById(c.Request().Context(), id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	plan, err := config.Server.Queries.GetStudyPlanById(c.Request().Context(), student.StudyPlanID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	room, err := config.Server.Queries.GetStudentRoomById(c.Request().Context(), student.RoomID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// validation based caching
	lastModified := student.UpdatedAt
	ETag := fmt.Sprintf("%x", sha256.Sum256([]byte(lastModified.Format(time.RFC3339))))

	modifiedSince := c.Request().Header.Get("If-Modified-Since")
	if c.Request().Header.Get("If-None-Match") == ETag || handler.IsLastModifiedValid(modifiedSince, lastModified) {
		return c.NoContent(http.StatusNotModified)
	}

	c.Response().Header().Set("ETag", ETag)
	c.Response().Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	c.Response().Header().Set("Cache-Control", "no-cache")

	return c.Render(http.StatusOK, "student-profile", Data{"Student": student, "Plan": plan, "Room": room})
}

func (config *webConfig) GetUpdateStudentPage(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	student, err := config.Server.Queries.GetStudentById(c.Request().Context(), id)
	if err != nil {
		return c.String(400, err.Error())
	}

	return c.Render(http.StatusOK, "update-student", Data{"Student": student})
}

func (config *webConfig) UpdateStudent(c echo.Context) error {
	time.Sleep(300 * time.Millisecond)
	ctx := c.Request().Context()
	type formParams struct {
		Email       string `validate:"email_constraints,cheeky_sql_inject"`
		PhoneNumber string `validate:"phone_constraints,cheeky_sql_inject"`
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	err = handler.WithTX(ctx, config.Server.DB, config.Server.Queries, func(qtx *database.Queries) error {
		params := formParams{
			Email:       c.FormValue("email"),
			PhoneNumber: c.FormValue("phone"),
		}

		if err := c.Validate(&params); err != nil {
			return err
		}

		_, err = qtx.UpdateStudent(ctx, database.UpdateStudentParams{
			ID:          id,
			Email:       params.Email,
			PhoneNumber: params.PhoneNumber,
			UpdatedAt:   time.Now(),
		})
		if err != nil {
			return err
		}

		// update updated_at for Last-Modified Header (caching)
		qtx.UpdateCollectionMetaLastModified(ctx, "student-coll")

		return nil
	})
	if err != nil {
		c.Render(http.StatusUnprocessableEntity, "update-student", Data{})
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": handler.SubmissionErrorMsg(err.Error()),
		})
	}

	redirectURL := fmt.Sprintf("/student/profile/%v", idStr)
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)
}
