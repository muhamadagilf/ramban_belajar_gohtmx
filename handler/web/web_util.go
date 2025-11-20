package web

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

var (
	MAJOR = []string{"TEKNIK INFORMATIKA", "REKAYASA PERANGKAT LUNAK", "AKUNTANSI"}
	ROOM  = []string{"TIR1", "TIR2", "RPLR1", "RPLR2", "AKR1", "AKR2"}
	YEAR  = time.Now().Year()
)

type Data = map[string]any

type webConfig struct {
	Server      *server.Server
	sessionName string
	store       *sessions.CookieStore
}

func studentsQueryParamHandler(c echo.Context, qtx *database.Queries) (Data, error) {
	var query struct {
		Search string `query:"search" validate:"omitempty,nochars,cheeky_sql_inject"`
		Room   string `query:"room" validate:"omitempty,oneof_room,cheeky_sql_inject"`
		Major  string `query:"major" validate:"omitempty,oneof_major,cheeky_sql_inject"`
	}

	rawURLQuery := c.Request().URL.RawQuery
	if err := c.Bind(&query); err != nil {
		return nil, err
	}

	if err := c.Validate(&query); err != nil {
		return nil, err
	}

	if rawURLQuery == "" {
		students, err := qtx.GetStudentAll(c.Request().Context())
		return Data{"Students": students}, err
	}

	if strings.Contains(rawURLQuery, "search") {
		if _, err := strconv.Atoi(query.Search); err != nil {
			students, err := qtx.GetStudentByNameOrNim(
				c.Request().Context(),
				database.GetStudentByNameOrNimParams{
					Name: "%" + strings.ToLower(query.Search) + "%",
					Nim:  "%%",
				})

			return Data{"Students": students}, err
		} else {
			students, err := qtx.GetStudentByNameOrNim(
				c.Request().Context(),
				database.GetStudentByNameOrNimParams{
					Name: "%%",
					Nim:  "%" + query.Search + "%",
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

func NewWebConfig() (*webConfig, error) {
	serverCfg, err := server.GetServerConfig()
	if err != nil {
		return nil, err
	}

	sessionKey := os.Getenv("session_key")
	if sessionKey == "" {
		return nil, errors.New("cannot find the sessionKey")
	}

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Domain:   "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	}

	return &webConfig{
		Server:      serverCfg,
		sessionName: "web_session",
		store:       store,
	}, nil
}

func IsCacheValid(c echo.Context, lastModified time.Time) (bool, string) {
	ETag := fmt.Sprintf("%x", sha256.Sum256([]byte(lastModified.Format(time.RFC3339))))
	modifiedSince := c.Request().Header.Get("If-Modified-Since")
	if c.Request().Header.Get("If-None-Match") == ETag || utils.IsLastModifiedValid(modifiedSince, lastModified) {
		return true, ETag
	}
	return false, ""
}
