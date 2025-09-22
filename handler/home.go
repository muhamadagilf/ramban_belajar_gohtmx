package handler

import "github.com/labstack/echo/v4"

func HomeHandler(c echo.Context) error {
	return c.Render(200, "index", Data{"Message": "Server Ready..."})
}
