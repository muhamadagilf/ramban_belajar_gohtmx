package web

import "github.com/labstack/echo/v4"

func (h *webConfig) GetHomePage(c echo.Context) error {
	return c.Render(200, "index", Data{"Message": "Server Ready..."})
}
