package web

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

func (config *webConfig) GetHomePage(c echo.Context) error {
	ctx := c.Request().Context()
	query := config.Server.Queries
	userID := c.Get("user_id").(uuid.UUID)

	user, err := query.GetUserById(ctx, userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	var userDat any

	switch user.Role {
	case handler.USER_ROLE_STUDENT:
		userDat, err = query.GetStudentByUserId(ctx, userID)
		if err != nil {
			return c.String(
				http.StatusInternalServerError,
				"Line 32: Internal Server Error",
			)
		}
	case handler.USER_ROLE_TEACHER:
		userDat = "Role: Teacher, comin soon"
	}

	return c.Render(http.StatusOK, "index", userDat)
}

func (config *webConfig) GetLoginPage(c echo.Context) error {
	log.Printf("\n\nLOGINPAGE\n%v\n\n", c.Request().Header.Get("Set-Cookie"))
	return c.Render(http.StatusOK, "login", Data{})
}

func (config *webConfig) LetUserLogin(c echo.Context) error {
	time.Sleep(300 * time.Millisecond)
	ctx := c.Request().Context()
	query := config.Server.Queries
	type formParams struct {
		Email    string `validate:"email_constraints,cheeky_sql_inject"`
		Password string `validate:"password_constraints"`
	}

	params := &formParams{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}

	// NOTE: validation and authentication
	if err := c.Validate(params); err != nil {
		return c.Render(http.StatusUnprocessableEntity, "login", Data{
			"Message": handler.ValidationErrorMsg(err.Error()),
		})
	}

	user, err := query.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return c.Render(http.StatusUnauthorized, "login", Data{
			"Message": handler.ERROR_FAILED_AUTHENTICATION,
		})
	}

	isUserValid := handler.CheckPasswordHash(params.Password, user.PasswordHash)
	if !isUserValid {
		return c.Render(http.StatusUnauthorized, "login", Data{
			"Message": handler.ERROR_FAILED_AUTHENTICATION,
		})
	}

	// NOTE: session creation
	sessionID := fmt.Sprintf("sess_id_%v_%v", user.ID, time.Now().Unix())
	session := sessions.NewSession(config.store, config.sessionName)
	session.Values["session_id"] = sessionID

	_, err = query.CreateUserSession(ctx, database.CreateUserSessionParams{
		SessionID: sessionID,
		UserID:    user.ID,
		ExpireAt:  time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		return c.Render(http.StatusInternalServerError, "login", Data{
			"Message": "Internal Server Error",
		})
	}

	// NOTE: save user session
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return c.Render(http.StatusInternalServerError, "login", Data{
			"Message": "Internal Server Error",
		})
	}

	log.Printf("\n\nLOGIN\n%v\n\n", c.Response().Header())

	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusOK)
}

func (config *webConfig) LetUserLogout(c echo.Context) error {
	time.Sleep(200 * time.Millisecond)
	ctx := c.Request().Context()
	query := config.Server.Queries

	session, err := config.store.Get(c.Request(), config.sessionName)
	if err != nil {
		c.Response().Header().Set("HX-Redirect", "/login")
		return c.NoContent(http.StatusOK)
	}

	sessionID, ok := session.Values["session_id"].(string)
	if ok && sessionID != "" {
		query.DeleteUserSession(ctx, sessionID)
	}

	// NOTE: cleans up the cooke, -1 to delete the cookie
	// so the browser dont have the user cookie anymore
	// session.Values = make(map[any]any)
	session.Options.MaxAge = -1
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	log.Printf("\n\nLOGOUT\n%v\n\n", c.Response().Header())

	c.Response().Header().Set("HX-Redirect", "/login")
	return c.NoContent(http.StatusOK)
}
