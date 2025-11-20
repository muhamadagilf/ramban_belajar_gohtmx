package web

import (
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

var skipperEndpoint = []string{
	"/login",
	"/admin/login",
}

func (config *webConfig) MiddlewareSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := config.store.Get(c.Request(), config.sessionName)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		c.Set("session", session)

		if err := session.Save(c.Request(), c.Response()); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return next(c)
	}
}

func (config *webConfig) MiddlewareAuthN(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		query := config.Server.Queries
		ctx := c.Request().Context()
		reqPath := c.Path()

		// skipper
		if slices.Contains(skipperEndpoint, reqPath) {
			session, err := config.store.Get(c.Request(), config.sessionName)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			sessionID, ok := session.Values["session_id"].(string)
			if sessionID == "" && !ok {
				log.Println("redirect cause no sessionID from cookie")
				return next(c)
			}

			_, err = query.GetUserSession(ctx, sessionID)
			if err != nil {
				log.Println("redirect cause no sessionID from DB")
				return next(c)
			}

			userIDStr, ok := session.Values["user_id"].(string)
			if !ok {
				return c.String(
					http.StatusInternalServerError,
					"Internal Server Error, at debug_block_auth:1",
				)
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.String(
					http.StatusInternalServerError,
					"Internal Server Error, at debug_block_auth:2",
				)
			}

			roles, err := query.GetUserRolesByUserID(ctx, userID)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			userRole := roles[0].Role

			switch len(roles) {
			case 1:
				if userRole == utils.USER_ROLE_STUDENT || userRole == utils.USER_ROLE_TEACHER {
					return c.Redirect(http.StatusFound, "/")
				} else {
					return c.Redirect(http.StatusFound, "/admin/panel")
				}
			case 2:
				return c.Redirect(http.StatusFound, "/")
			}

		}

		// If request un-skipper endpoint goes right up here
		session, err := config.store.Get(c.Request(), config.sessionName)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		sessionID, ok := session.Values["session_id"].(string)
		if !ok && sessionID == "" {
			return c.Redirect(http.StatusFound, "/login")
		}

		sessionDat, err := query.GetUserSession(ctx, sessionID)
		if err != nil || sessionDat.IsRevoked || time.Now().After(sessionDat.ExpireAt) {
			query.DeleteUserSession(ctx, sessionID)

			session.Options.MaxAge = -1
			if err := session.Save(c.Request(), c.Response()); err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			return c.Redirect(http.StatusFound, "/login")
		}

		// update last_activity, everytime user make a request
		if err := query.UpdateLastActivityUserSession(ctx, sessionID); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		c.Set("user_id", sessionDat.UserID)

		return next(c)
	}
}
