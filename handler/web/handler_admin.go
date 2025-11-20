package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

func (config *webConfig) GetAdminLoginPage(c echo.Context) error {
	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, at debug_block:1",
		)
	}

	return c.Render(http.StatusOK, "login-page", Data{
		"Role":       utils.USER_ROLE_ADMIN,
		"CSRF_Token": CSRFToken,
	})
}

func (config *webConfig) GetAdminPanelPage(c echo.Context) error {
	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, at debug_block:1",
		)
	}

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code: ERR00500",
		)
	}

	if allowed, _ := config.Server.Can(claims, "adminPanelPages", "view"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{
			"Message": utils.ERROR_USER_UNAUTHORIZED,
		})
	}

	return c.Render(http.StatusOK, "home", Data{
		"CSRF_Token": CSRFToken,
		"UserRole":   claims.Roles[0],
	})
}
