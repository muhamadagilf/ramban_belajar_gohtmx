package web

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

func (config *webConfig) GetHomePage(c echo.Context) error {
	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code: ERR103500",
		)
	}

	allowed, role := config.Server.Can(claims, "homePage", "view")
	if !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{})
	}

	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code: ERR101500",
		)
	}

	c.Response().Header().Set("Cache-Control", "max-age=3600, private")
	return c.Render(http.StatusOK, "home", Data{
		"CSRF_Token": CSRFToken,
		"UserID":     claims.UserID,
		"UserRole":   role,
	})
}

func (config *webConfig) GetLoginPage(c echo.Context) error {
	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, at debug_block:1",
		)
	}

	return c.Render(http.StatusOK, "login-page", Data{
		"Role":       utils.USER_ROLE_STUDENT,
		"CSRF_Token": CSRFToken,
	})
}

func (config *webConfig) Login(redirectURL string, role string) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		query := config.Server.Queries
		CSRFToken, ok := c.Get("csrf").(string)
		if !ok {
			log.Println("Internal Server Error at code:ERR78500")
			return c.String(
				http.StatusInternalServerError,
				"Internal Server Error, Contact Support with code:ERR78500",
			)
		}

		type formParams struct {
			Email    string `validate:"email_constraints,cheeky_sql_inject"`
			Password string `validate:"password_constraints"`
		}

		params := &formParams{
			Email:    c.FormValue("email"),
			Password: c.FormValue("password"),
		}

		// validation
		if err := c.Validate(params); err != nil {
			return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
				"Message":    utils.ValidationErrorMsg(err.Error()),
				"Role":       role,
				"CSRF_Token": CSRFToken,
			})
		}

		user, err := query.GetUserByEmail(ctx, params.Email)
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		userRoles, err := config.Server.LoadUserRoles(ctx, user.ID)
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		switch len(userRoles) {
		case 1:
			if !slices.Contains(userRoles, role) && userRoles[0] == utils.USER_ROLE_ADMIN {
				c.Response().Header().Set("HX-Redirect", "/admin/login")
				return c.NoContent(http.StatusOK)
			}

			if !slices.Contains(userRoles, role) && userRoles[0] != utils.USER_ROLE_ADMIN {
				c.Response().Header().Set("HX-Redirect", "/login")
				return c.NoContent(http.StatusOK)
			}
		}

		isUserValid := utils.CheckPasswordHash(params.Password, user.PasswordHash)
		if !isUserValid {
			return c.Render(http.StatusUnauthorized, "error-message", Data{
				"Message":    utils.ERROR_FAILED_AUTHENTICATION,
				"Role":       role,
				"CSRF_Token": CSRFToken,
			})
		}

		// session creation
		sessionID := fmt.Sprintf("sess_id_%v_%v", user.ID, time.Now().Unix())
		session, err := config.store.Get(c.Request(), config.sessionName)
		if err != nil {
			log.Println(err)
			return c.String(
				http.StatusInternalServerError,
				fmt.Sprintf("Internal Server Error, %v", err.Error()),
			)
		}

		session.Values["session_id"] = sessionID
		session.Values["user_id"] = fmt.Sprintf("%v", user.ID)

		_, err = query.CreateUserSession(ctx, database.CreateUserSessionParams{
			SessionID: sessionID,
			UserID:    user.ID,
			ExpireAt:  time.Now().Add(24 * time.Hour),
		})
		if err != nil {
			log.Println(err)
			return c.String(
				http.StatusInternalServerError,
				fmt.Sprintf("Internal Server Error, %v", err.Error()),
			)
		}

		if err := session.Save(c.Request(), c.Response()); err != nil {
			log.Println(err)
			return c.String(
				http.StatusInternalServerError,
				fmt.Sprintf("Internal Server Error, %v", err.Error()),
			)
		}

		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
}

func (config *webConfig) Logout(redirectURL string) echo.HandlerFunc {
	return func(c echo.Context) error {
		time.Sleep(200 * time.Millisecond)
		ctx := c.Request().Context()
		query := config.Server.Queries

		session, err := config.store.Get(c.Request(), config.sessionName)
		if err != nil {
			return c.String(
				http.StatusInternalServerError,
				fmt.Sprintf("Internal Server Error, %v", err.Error()),
			)
		}

		sessionID, ok := session.Values["session_id"].(string)
		if ok && sessionID != "" {
			if err := query.DeleteUserSession(ctx, sessionID); err != nil {
				return c.String(
					http.StatusInternalServerError,
					fmt.Sprintf("Internal Server Error, %x", err),
				)
			}
		}

		// cleans up the cooke, -1 to delete the cookie
		// so the browser dont have the user cookie anymore
		session.Options.MaxAge = -1
		if err := session.Save(c.Request(), c.Response()); err != nil {
			return c.String(
				http.StatusInternalServerError,
				fmt.Sprintf("Internal Server Error, %x", err),
			)
		}

		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
}
