package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler"
)

func (config *webConfig) GetHomePage(c echo.Context) error {
	return c.Render(http.StatusOK, "index", Data{"Message": "Server Ready..."})
}

func (config *webConfig) GetLoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login", Data{})
}

func (config *webConfig) LetUserLogin(c echo.Context) error {
	ctx := c.Request().Context()
	query := config.Server.Queries
	type formParams struct {
		Email    string `validate:"email_constraints,cheeky_sql_inject"`
		Password string
	}

	params := &formParams{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}

	if err := c.Validate(params); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// authentication here
	// and check the password, if valid let user login
	// if not punch ERROR_USER_UNAUTHENTICATED
	hash, err := query.GetUserHash(ctx, params.Email)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	isUserValid := handler.CheckPasswordHash(params.Password, hash)
	if !isUserValid {
		return c.String(http.StatusBadRequest, handler.ERROR_USER_UNAUTHENTICATED)
	}

	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusOK)
}
