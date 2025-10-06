package app

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
)

var MAJOR = []string{"TEKNIK INFORMATIKA", "REKAYASA PERANGKAT LUNAK", "AKUNTANSI"}
var ROOM = []string{"TIR1", "TIR2", "RPLR1", "RPLR2", "AKR1", "AKR2"}
var YEAR = time.Now().Year()

type appConfig struct {
	Server *server.Server
}

func submissionErrorMsg(err string) string {
	err = strings.ToLower(err)
	if strings.Contains(err, `violates check constraint "students_phone_number_check"`) {
		return "error: invalid phone number, please input the correct number"
	}

	if strings.Contains(err, `violates unique constraint "students_phone_number_key"`) {
		return "error: cannot input phone number, number already exists"
	}

	if strings.Contains(err, `violates unique constraint "students_email_key"`) {
		return "error: cannot input email, email already exists"
	}

	if strings.Contains(err, `violates unique constraint "students_nip_key"`) {
		return "error: cannot input NIP, NIP already exists"
	}

	return err
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func studentsQueryParamHandler(c echo.Context, qtx *database.Queries) (Data, error) {
	type StudentsQuery struct {
		Search string `query:"search" validate:"omitempty,nochars"`
		Room   string `query:"room" validate:"omitempty,oneof_room"`
		Major  string `query:"major" validate:"omitempty,oneof_major"`
	}

	rawUrlQuery := c.Request().URL.RawQuery
	query := StudentsQuery{}
	if err := c.Bind(&query); err != nil {
		return Data{}, err
	}

	if err := c.Validate(&query); err != nil {
		return Data{}, err
	}

	if rawUrlQuery == "" {
		students, err := qtx.GetStudentAll(c.Request().Context())
		return Data{"Students": students}, err
	}

	if strings.Contains(rawUrlQuery, "search") {
		if _, err := strconv.Atoi(query.Search); err != nil {
			students, err := qtx.GetStudentByNameOrNim(
				c.Request().Context(),
				database.GetStudentByNameOrNimParams{
					Name: "%" + query.Search + "%",
					Nim:  "%%",
				})

			return Data{"Students": students}, err
		} else {
			students, err := qtx.GetStudentByNameOrNim(
				c.Request().Context(),
				database.GetStudentByNameOrNimParams{
					Name: "%" + query.Search + "%",
					Nim:  "%%",
				})

			return Data{"Students": students}, err
		}
	}

	students, err := qtx.GetStudentsByRoomAndMajor(
		c.Request().Context(),
		database.GetStudentsByRoomAndMajorParams{
			Name:  query.Room,
			Major: query.Major,
		})

	return Data{"Students": students}, err
}

func NewAppConfig() (*appConfig, error) {
	server, err := server.GetServerConfig()
	if err != nil {
		return nil, err
	}

	return &appConfig{
		Server: server,
	}, nil
}
