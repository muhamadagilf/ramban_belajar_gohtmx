package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
)

var ERROR_CLASS_FULL = "cannot assign to the major, due to the full student"
var ERROR_INVALID_NIP = "error: invalid nomer induk pengguna (nip), please check your birthdate/nip"

type apiConfig struct {
	Server *server.Server
}

type studentData struct {
	StudyPlan database.StudyPlan
	Room      database.Room
}

func NewApiConfig() (*apiConfig, error) {
	server, err := server.GetServerConfig()
	if err != nil {
		return nil, err
	}

	return &apiConfig{
		Server: server,
	}, nil
}

func (config *apiConfig) HandlerMiddlewareStudent(next echo.HandlerFunc) echo.HandlerFunc {
	var major struct {
		Major string `json:"major"`
	}

	return func(c echo.Context) error {
		ctx := c.Request().Context()
		q := config.Server.Queries

		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Data{
				"error": err.Error(),
				"hint":  "here daddy, line 50",
			})
		}

		c.Request().Body.Close()
		c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err = json.Unmarshal(bodyBytes, &major); err != nil {
			return c.JSON(http.StatusBadRequest, Data{
				"error": err.Error(),
				"hint":  "here daddy, line 60",
			})
		}

		var roomPrefix string
		switch major.Major {
		case "TEKNIK INFORMATIKA":
			roomPrefix = "TI"
		case "REKAYASA PERANGKAT LUNAK":
			roomPrefix = "RPL"
		case "AKUNTANSI":
			roomPrefix = "AK"
		}

		studyPlan, err := q.GetStudyPlan(ctx, database.GetStudyPlanParams{
			Semester: int32(1),
			Major:    major.Major,
		})

		if err != nil {
			return c.JSON(http.StatusBadRequest, Data{
				"error": err.Error(),
				"hint":  "here daddy, line 82",
			})
		}

		pattern := "%" + roomPrefix + "%"
		rooms, err := q.GetStudentRoom(ctx, pattern)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Data{
				"error": err.Error(),
				"hint":  "here daddy, line 91",
			})
		}

		var room database.Room
		studentClassCount := major.Major + "-StudentCount"
		studentCount, _ := q.GetCollectionMetaValue(ctx, studentClassCount)
		n, _ := strconv.Atoi(studentCount)
		if n > 10 {
			return c.JSON(http.StatusBadRequest, Data{"error": ERROR_CLASS_FULL})
		}
		if n < 5 {
			room = rooms[0]
		}
		if n >= 5 && n < 11 {
			room = rooms[1]
		}

		c.Set("studentInfo", &studentData{StudyPlan: studyPlan, Room: room})

		return next(c)
	}
}
