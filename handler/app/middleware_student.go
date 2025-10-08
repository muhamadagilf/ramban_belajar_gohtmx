package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

type StudentData struct {
	StudyPlan database.StudyPlan
	Room      database.Room
}

func (config *webConfig) MiddlewareStudent(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var roomPrefix string
		ctx := c.Request().Context()
		qtx := config.Server.Queries

		major := c.FormValue("major")

		studyPlan, err := qtx.GetStudyPlan(ctx, database.GetStudyPlanParams{
			Semester: int32(1),
			Major:    major,
		})

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		switch major {
		case "TEKNIK INFROMATIKA":
			roomPrefix = "TI"
		case "REKAYASA PERANGKAT LUNAK":
			roomPrefix = "RPL"
		case "AKUNTANSI":
			roomPrefix = "AK"
		}

		pattern := "%" + roomPrefix + "%"
		var room database.Room

		rooms, err := qtx.GetStudentRoom(ctx, pattern)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		if len(rooms) == 0 {
			c.String(http.StatusBadRequest, err.Error())
		}

		studentClassCount := major + "-StudentCount"
		studentCount, err := qtx.GetCollectionMetaValue(ctx, studentClassCount)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		n, _ := strconv.Atoi(studentCount)
		if n > 10 {
			return c.String(
				http.StatusBadRequest,
				fmt.Sprintf("Class from %v is all full", major),
			)
		}
		if n < 5 {
			room = rooms[0]
		}
		if n >= 5 && n < 11 {
			room = rooms[1]
		}
		c.Set("studentData", &StudentData{StudyPlan: studyPlan, Room: room})

		return next(c)

	}
}
