// Package utils
package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var LoginLimiter = middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
	Skipper: middleware.DefaultSkipper,
	Store:   middleware.NewRateLimiterMemoryStore(5),
	IdentifierExtractor: func(c echo.Context) (string, error) {
		return c.RealIP(), nil
	},
	ErrorHandler: func(c echo.Context, err error) error {
		return c.String(http.StatusTooManyRequests, "Too many attempts. Try again later")
	},
	DenyHandler: func(c echo.Context, identifier string, err error) error {
		return c.String(http.StatusTooManyRequests, "Rate limit exceeded")
	},
})
