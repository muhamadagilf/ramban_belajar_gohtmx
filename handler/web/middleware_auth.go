package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// func IsUserAuthenticated(request *http.Request, config *webConfig) (bool, error) {
// 	session, err := config.store.Get(request, config.sessionName)
// 	if err != nil {
// 		return false, err
// 	}
// 	_, ok := session.Values["session_id"].(string)
// 	return ok, nil
// }

func (config *webConfig) MiddlewareAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := config.Server.Queries
		ctx := c.Request().Context()

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

		return next(c)
	}
}
