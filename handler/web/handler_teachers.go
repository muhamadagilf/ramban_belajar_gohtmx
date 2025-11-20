package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

func (config *webConfig) GetCoursePage(c echo.Context) error {
	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(http.StatusInternalServerError, "Please Contact Support with code:76500")
	}

	if allowed, _ := config.Server.Can(claims, "coursePage", "view"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{})
	}

	return c.Render(http.StatusOK, "course-page", Data{})
}

func (config *webConfig) CreateCourse(c echo.Context) error {
	context := c.Request().Context()
	query := config.Server.Queries
	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			utils.InternalServerErrorMessage("ERR74500", ""),
		)
	}

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			utils.InternalServerErrorMessage("ERR44500", ""),
		)
	}

	if allowed, _ := config.Server.Can(claims, "courses", "create"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{})
	}

	type formParams struct {
		Title     string
		Desc      string
		CreatedAt string
	}

	params := &formParams{
		Title:     c.FormValue("course_title"),
		Desc:      c.FormValue("course_desc"),
		CreatedAt: c.FormValue("course_date"),
	}

	if err := c.Validate(params); err != nil {
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"Message": utils.ERROR_INVALID_INPUT_DATA,
		})
	}

	// Generate ID to identified the file from storage
	// Store it to the DB with corresponds course entry
	coursesStoragePath := os.Getenv("course_storage_path")
	courseFileID := fmt.Sprintf("%s_id_%v", params.Title, time.Now().Unix())

	file, err := c.FormFile("course")
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			utils.InternalServerErrorMessage("ERR88500", err.Error()),
		)
	}

	defer c.Request().MultipartForm.RemoveAll()

	src, err := file.Open()
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			utils.InternalServerErrorMessage("ERR87500", err.Error()),
		)
	}

	defer src.Close()

	dst, err := os.Create(coursesStoragePath + courseFileID)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			utils.InternalServerErrorMessage("ERR89500", err.Error()),
		)
	}

	defer dst.Close()

	// COPIES TO STAGING STORAGE (temp)
	if _, err = io.Copy(dst, src); err != nil {
		os.Remove(coursesStoragePath + courseFileID)
		return c.String(
			http.StatusInternalServerError,
			utils.InternalServerErrorMessage("ERR99500", err.Error()),
		)
	}

	if err = utils.WithTX(context, config.Server.DB, query, func(qtx *database.Queries) error {
		return nil
	}); err != nil {
		os.Remove(coursesStoragePath + courseFileID)
		return c.Render(http.StatusInternalServerError, "error-message", Data{
			"CSRF_Token": CSRFToken,
		})
	}

	// Send Message to Queue to process moves the file to permanent storge (validation file involved)

	c.Response().Header().Set("HX-Redirect", "/courses/create")
	return c.NoContent(http.StatusCreated)
}
