package handler

import (
	"context"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

type StudentData struct {
	StudyPlan database.StudyPlan
	Room      database.Room
}

func (srv *Server) MiddlewareStudent(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		var roomPrefix string

		major := c.FormValue("major")

		studyPlan, err := srv.Queries.GetStudyPlan(context.Background(), database.GetStudyPlanParams{
			Semester: int32(1),
			Major:    major,
		})

		if err != nil {
			log.Println("errror: couldnt get the study plan")
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

		rooms, err := srv.Queries.GetStudentRoom(context.Background(), pattern)
		if err != nil {
			log.Println(err.Error())
		}

		if len(rooms) == 0 {
			log.Println(err.Error())
		}

		c.Set("studentData", &StudentData{StudyPlan: studyPlan, Room: rooms[0]})

		return next(c)

	}
}
