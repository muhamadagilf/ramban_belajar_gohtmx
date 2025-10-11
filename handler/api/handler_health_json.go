// Package api
package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Data = map[string]any

func (config *apiConfig) HandlerHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, Data{"Message": "from :3000 up and running..."})
}
