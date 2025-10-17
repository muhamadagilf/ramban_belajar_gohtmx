package web

import (
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var skipperEndpoint = []string{
	"/login",
	"/students/submission",
}

func (config *webConfig) MiddlewareAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := config.Server.Queries
		ctx := c.Request().Context()

		if slices.Contains(skipperEndpoint, c.Path()) {
			session, err := config.store.Get(c.Request(), config.sessionName)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Internal Server Error")
			}

			sessionID, ok := session.Values["session_id"].(string)
			if ok && sessionID != "" {
				sessionDat, err := query.GetUserSession(ctx, sessionID)
				if err == nil && sessionDat.UserID != uuid.Nil {
					return c.Redirect(http.StatusFound, "/")
				}
			}

			c.Set("session", session)

			if err := session.Save(c.Request(), c.Response()); err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			return next(c)

		}

		session, err := config.store.Get(c.Request(), config.sessionName)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		sessionID, ok := session.Values["session_id"].(string)
		if !ok {
			return c.Redirect(http.StatusFound, "/login")
		}

		sessionDat, err := query.GetUserSession(ctx, sessionID)
		if err != nil || sessionDat.IsRevoked || time.Now().After(sessionDat.ExpireAt) {
			query.DeleteUserSession(ctx, sessionID)
			return c.Redirect(http.StatusFound, "/login")
		}

		// NOTE: update last_activity if session valid
		query.UpdateLastActivityUserSession(ctx, sessionID)

		c.Set("session_id", sessionID)
		c.Set("user_id", sessionDat.UserID)
		c.Set("session", session)

		return next(c)
	}
}
