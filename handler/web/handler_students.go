// Package web
package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

func (config *webConfig) GetStudentSubmitPage(c echo.Context) error {
	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code:ERR21500",
		)
	}

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code:ERR22500",
		)
	}

	if allowed, _ := config.Server.Can(claims, "studentCreatePage", "view"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{
			"Message": utils.ERROR_USER_UNAUTHORIZED,
		})
	}

	return c.Render(http.StatusOK, "student-submission", Data{
		"Major":      utils.MAJOR,
		"CSRF_Token": CSRFToken,
	})
}

func (config *webConfig) CreateStudent(c echo.Context) error {
	time.Sleep(200 * time.Millisecond)
	ctx := c.Request().Context()

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code:ERR22500",
		)
	}

	if allowed, _ := config.Server.Can(claims, "students", "create"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{
			"Message": utils.ERROR_USER_UNAUTHORIZED,
		})
	}

	type formParams struct {
		Name            string `validate:"name_constraints,cheeky_sql_inject"`
		Email           string `validate:"email_constraints,cheeky_sql_inject"`
		PhoneNumber     string `validate:"phone_constraints"`
		Nip             string `validate:"nip_constraints"`
		DateOfBirth     string `validate:"cheeky_sql_inject"`
		Password        string `validate:"password_constraints"`
		ConfirmPassword string `validate:"password_constraints"`
	}

	err := utils.WithTX(ctx, config.Server.DB, config.Server.Queries, func(qtx *database.Queries) error {
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
			return errors.New(utils.ERROR_INVALID_CONFIRM_PASSWORD)
		}

		if !utils.IsNIPValid(params.Nip, params.DateOfBirth) {
			return errors.New(utils.ERROR_INVALID_NIP)
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
		hashedPassword, err := utils.HashPassword(params.Password)
		if err != nil {
			return err
		}

		// doing user & student creation
		user, err := qtx.CreateUser(ctx, database.CreateUserParams{
			Email:        params.Email,
			PasswordHash: hashedPassword,
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
			Year:        int32(time.Now().Year()),
			StudyPlanID: studentDat.StudyPlan.ID,
			RoomID:      studentDat.Room.ID,
			UserID:      user.ID,
		})
		if err != nil {
			return err
		}

		// assign student role
		_, err = qtx.CreateUserRoles(ctx, database.CreateUserRolesParams{
			UserID: user.ID,
			Role:   utils.USER_ROLE_STUDENT,
		})
		if err != nil {
			return err
		}

		// add the student to the classroom
		if err = qtx.SetStudentClassroom(ctx, database.SetStudentClassroomParams{
			RoomID:    student.RoomID,
			StudentID: student.ID,
		}); err != nil {
			return nil
		}

		// updated the collection_meta "-StudentCount", increment the count
		// to keep track of the member of class
		studentClassCount := studentDat.StudyPlan.Major + "-StudentCount"
		if err = qtx.IncrementValueByname(ctx, studentClassCount); err != nil {
			return err
		}

		// update updated_at for Last-Modified Header (caching)
		if err = qtx.UpdateCollectionMetaLastModified(ctx, "student-coll"); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		c.Render(http.StatusUnprocessableEntity, "student-submission", Data{})
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": utils.ValidationErrorMsg(err.Error()),
		})
	}

	c.Response().Header().Set("HX-Redirect", "/login")
	return c.NoContent(http.StatusCreated)
}

func (config *webConfig) GetStudentProfile(c echo.Context) error {
	ctx := c.Request().Context()
	query := config.Server.Queries

	IDStr := c.Param("id")
	paramUserID, err := uuid.Parse(IDStr)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code:ERR44500",
		)
	}

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code:ERR22500",
		)
	}

	switch claims.Roles[0] {
	case utils.USER_ROLE_STUDENT:
		if claims.UserID != paramUserID {
			return c.Render(http.StatusUnauthorized, "unauthorized", Data{})
		}
	}

	if allowed, _ := config.Server.Can(claims, "students", "view"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{})
	}

	student, err := query.GetStudentByUserId(ctx, claims.UserID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	plan, err := query.GetStudyPlanById(ctx, student.StudyPlanID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	room, err := query.GetStudentRoomById(ctx, student.RoomID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// validation based caching
	lastModified := student.UpdatedAt

	valid, ETag := IsCacheValid(c, lastModified)
	if valid {
		return c.NoContent(http.StatusNotModified)
	}

	c.Response().Header().Set("ETag", ETag)
	c.Response().Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	c.Response().Header().Set("Cache-Control", "no-cache")
	return c.Render(http.StatusOK, "student-profile", Data{
		"Student":  student,
		"Plan":     plan,
		"Room":     room,
		"UserRole": claims.Roles[0],
	})
}

func (config *webConfig) GetUpdateStudentPage(c echo.Context) error {
	ctx := c.Request().Context()
	query := config.Server.Queries
	userID := c.Get("user_id").(uuid.UUID)
	CRSFToken := c.Get("csrf").(string)

	student, err := query.GetStudentByUserId(ctx, userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "update-student", Data{
		"Student":    student,
		"CSRF_Token": CRSFToken,
	})
}

func (config *webConfig) UpdateStudent(c echo.Context) error {
	time.Sleep(200 * time.Millisecond)
	ctx := c.Request().Context()
	query := config.Server.Queries
	userID := c.Get("user_id").(uuid.UUID)
	type formParams struct {
		Email       string `validate:"email_constraints,cheeky_sql_inject"`
		PhoneNumber string `validate:"phone_constraints,cheeky_sql_inject"`
	}

	student, err := query.GetStudentByUserId(ctx, userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	err = utils.WithTX(ctx, config.Server.DB, query, func(qtx *database.Queries) error {
		params := formParams{
			Email:       c.FormValue("email"),
			PhoneNumber: c.FormValue("phone"),
		}

		if err := c.Validate(&params); err != nil {
			return err
		}

		_, err = qtx.UpdateStudent(ctx, database.UpdateStudentParams{
			ID:          student.ID,
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
			"Message": utils.ValidationErrorMsg(err.Error()),
		})
	}

	redirectURL := fmt.Sprintf("/students/%v/profile", student.ID)
	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)
}

// NOTE: admin level utilsFunc

func (config *webConfig) GetStudentsPage(c echo.Context) error {
	ctx := c.Request().Context()
	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, at debug_block_getstudents:1",
		)
	}

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code:ERR20500",
		)
	}

	if allowed, _ := config.Server.Can(claims, "adminPanelPages", "view"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{
			"Message": utils.ERROR_USER_UNAUTHORIZED,
		})
	}

	// retrieves all the necessary data, including query params handling
	// do some filter & search querying
	studentsPageData, err := studentsQueryParamHandler(c, config.Server.Queries)
	if err != nil {
		return c.String(http.StatusUnprocessableEntity, err.Error())
	}

	studentsPageData["Rooms"] = utils.ROOM
	studentsPageData["Majors"] = utils.MAJOR
	studentsPageData["CSRF_Token"] = CSRFToken
	studentsPageData["UserRole"] = claims.Roles[0]

	// do validation based caching
	lastModified, err := config.Server.Queries.GetCollectionMetaLastModified(ctx, "student-coll")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	valid, ETag := IsCacheValid(c, lastModified)
	if valid {
		return c.NoContent(http.StatusNotModified)
	}

	c.Response().Header().Set("ETag", ETag)
	c.Response().Header().Set("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
	c.Response().Header().Set("Cache-Control", "no-cache")
	return c.Render(http.StatusOK, "db-students-panel", studentsPageData)
}

func (config *webConfig) DeleteStudent(c echo.Context) error {
	time.Sleep(300 * time.Millisecond)
	ctx := c.Request().Context()
	err := utils.WithTX(ctx, config.Server.DB, config.Server.Queries, func(qtx *database.Queries) error {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		student, err := qtx.DeleteStudentById(ctx, id)
		if err != nil {
			return err
		}

		err = qtx.DeleteUserByID(ctx, student.UserID)
		if err != nil {
			return err
		}

		// updated the collection_meta "-StudentCount"
		// decrement the student count, if there is deletion
		studentPlan, err := qtx.GetStudyPlanById(ctx, student.StudyPlanID)
		if err != nil {
			return err
		}

		studentClassCount := studentPlan.Major + "-StudentCount"
		if err = qtx.DecrementValueByName(ctx, studentClassCount); err != nil {
			return err
		}

		// simply, add the nim to freelist, if there is student deletion
		if err = qtx.AddToFreelist(ctx, student.Nim); err != nil {
			return err
		}

		// update updated_at for Last-Modified Header (caching)
		if err = qtx.UpdateCollectionMetaLastModified(ctx, "student-coll"); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": err.Error(),
		})
	}

	c.Response().Header().Set("HX-Redirect", "/admin/panel/students")
	return c.NoContent(http.StatusOK)
}
