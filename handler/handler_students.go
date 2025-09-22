package handler

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

type Data map[string]interface{}

func (srv *Server) GetStudentSubmitPage(c echo.Context) error {
	return c.Render(http.StatusOK, "student-submission", Data{
		"Major": MAJOR,
	})
}

func (srv *Server) CreateStudent(c echo.Context) error {
	ctx := c.Request().Context()
	err := WithTX(ctx, srv.DB, srv.Queries, func(qtx *database.Queries) error {
		name := c.FormValue("name")
		email := c.FormValue("email")
		phone := c.FormValue("phone")

		if name == " " || phone == " " {
			return fmt.Errorf("error: Required name, phone and email")
		}

		nipStr := c.FormValue("nip")
		nip, err := strconv.ParseInt(nipStr, 10, 64)
		if err != nil {
			return fmt.Errorf("error: invalid NIP")
		}

		nim, err := generateNIM(qtx)
		if err != nil {
			return err
		}

		studentDat := c.Get("studentData").(*StudentData)

		student, err := qtx.CreateStudent(ctx, database.CreateStudentParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Nim:         nim,
			Nip:         int32(nip),
			Name:        name,
			Email:       email,
			PhoneNumber: phone,
			Year:        int32(YEAR),
			StudyPlanID: studentDat.StudyPlan.ID,
			RoomID:      studentDat.Room.ID,
		})

		if err != nil {
			return err
		}

		c.Set("createdStudent", &student)

		err = qtx.SetStudentClassroom(ctx, database.SetStudentClassroomParams{
			RoomID:    student.RoomID,
			StudentID: student.ID,
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.Render(http.StatusUnprocessableEntity, "student-submission", Data{})
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": submissionErrorMsg(err.Error()),
		})
	}

	createdStudent := c.Get("createdStudent").(*database.Student)
	moreStudentInfo := c.Get("studentData").(*StudentData)

	c.Render(http.StatusCreated, "student-submission", Data{})
	return c.Render(http.StatusCreated, "submission-succeed", Data{
		"Student":         createdStudent,
		"MoreStudentInfo": moreStudentInfo,
		"Message":         "Submission Complete",
	})
}

func (srv *Server) GetStudentsPage(c echo.Context) error {
	students, err := srv.Queries.GetStudentAll(c.Request().Context())
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	plans := []database.StudyPlan{}

	for _, student := range students {
		plan, err := srv.Queries.GetStudyPlanById(c.Request().Context(), student.StudyPlanID)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		plans = append(plans, plan)
	}

	version := students[0].UpdatedAt.Format(time.RFC3339)
	lastModified := students[0].UpdatedAt
	eTag := fmt.Sprintf("%x", sha256.Sum256([]byte(version)))

	modifiedSince := c.Request().Header.Get("If-Modified-Since")
	if c.Request().Header.Get("If-None-Match") == eTag || isLastModifiedValid(modifiedSince, lastModified) {
		return c.NoContent(http.StatusNotModified)
	}

	c.Response().Header().Set("ETag", eTag)
	c.Response().Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	c.Response().Header().Set("Cache-Control", "no-cache")

	return c.Render(http.StatusOK, "students", Data{"Students": students, "Plans": plans})
}

func (srv *Server) DeleteStudent(c echo.Context) error {
	time.Sleep(500 * time.Millisecond)
	ctx := c.Request().Context()
	err := WithTX(ctx, srv.DB, srv.Queries, func(qtx *database.Queries) error {

		idStr := c.Param("id")

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		err = qtx.DeleteStudentById(ctx, id)
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

	return c.NoContent(http.StatusOK)
}

func (srv *Server) GetUpdateStudentPage(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	student, err := srv.Queries.GetStudentById(c.Request().Context(), id)
	if err != nil {
		return c.String(400, err.Error())
	}

	return c.Render(http.StatusOK, "update-student", Data{"Student": student})
}

func (srv *Server) UpdateStudent(c echo.Context) error {
	time.Sleep(500 * time.Millisecond)
	ctx := c.Request().Context()
	err := WithTX(ctx, srv.DB, srv.Queries, func(qtx *database.Queries) error {

		idStr := c.Param("id")

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		email := c.FormValue("email")
		phone := c.FormValue("phone")

		if email == " " || phone == " " {
			return fmt.Errorf("error: Required email and phone")
		}

		_, err = qtx.UpdateStudent(ctx, database.UpdateStudentParams{
			ID:          id,
			Email:       email,
			PhoneNumber: phone,
			UpdatedAt:   time.Now(),
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.Render(http.StatusUnprocessableEntity, "update-student", Data{})
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": submissionErrorMsg(err.Error()),
		})
	}

	c.Response().Header().Set("HX-Redirect", "/students")
	return c.NoContent(http.StatusOK)
}
