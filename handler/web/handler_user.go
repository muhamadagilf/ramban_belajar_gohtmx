package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

func (config *webConfig) GetUsersPage(c echo.Context) error {
	context := c.Request().Context()
	query := config.Server.Queries

	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, at debug_block_getusers:1",
		)
	}

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code: ERR10500",
		)
	}

	if allowed, _ := config.Server.Can(claims, "adminPanelPages", "view"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{
			"Message": utils.ERROR_USER_UNAUTHORIZED,
		})
	}

	users, err := query.GetUsersAllJoinRoles(context)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, at debug_block_getusers:2",
		)
	}

	return c.Render(http.StatusOK, "db-users-panel", Data{
		"CSRF_Token": CSRFToken,
		"Users":      users,
		"UserRole":   claims.Roles[0],
	})
}

func (config *webConfig) CreateUser(c echo.Context) error {
	time.Sleep(200 * time.Millisecond)
	ctx := c.Request().Context()
	query := config.Server.Queries

	claims, ok := c.Get("claims").(*server.Claims)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, Contact Support with code: ERR10500",
		)
	}

	if allowed, _ := config.Server.Can(claims, "users", "create"); !allowed {
		return c.Render(http.StatusUnauthorized, "unauthorized", Data{
			"Message": utils.ERROR_USER_UNAUTHORIZED,
		})
	}

	CSRFToken, ok := c.Get("csrf").(string)
	if !ok {
		return c.String(
			http.StatusInternalServerError,
			"Internal Server Error, at debug_block_createusers:1",
		)
	}

	type formParams struct {
		Email    string `validate:"email_constraints,cheeky_sql_inject"`
		Password string `validate:"password_constraints"`
		Role     string `validate:"roles_checks"`
	}

	params := &formParams{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
		Role:     c.FormValue("roles"),
	}

	err := utils.WithTX(ctx, config.Server.DB, query, func(qtx *database.Queries) error {
		if err := c.Validate(params); err != nil {
			return err
		}

		passwordHashed, err := utils.HashPassword(params.Password)
		if err != nil {
			return err
		}

		user, err := qtx.CreateUser(ctx, database.CreateUserParams{
			Email:        params.Email,
			PasswordHash: passwordHashed,
		})
		if err != nil {
			return err
		}

		// if role "superuser", assign with two user-type
		// performance-wise might be bad (iterate those user-type)
		switch params.Role {
		case utils.USER_ROLE_SUPERUSER:
			for _, role := range []string{
				utils.USER_ROLE_ADMIN,
				utils.USER_ROLE_TEACHER,
			} {
				_, err = qtx.CreateUserRoles(ctx, database.CreateUserRolesParams{
					UserID: user.ID,
					Role:   role,
				})
				if err != nil {
					return err
				}
			}
		default:
			_, err = qtx.CreateUserRoles(ctx, database.CreateUserRolesParams{
				UserID: user.ID,
				Role:   params.Role,
			})
			if err != nil {
				return err
			}

		}

		return nil
	})
	if err != nil {
		return c.Render(http.StatusUnprocessableEntity, "error-message", Data{
			"CSRF_Token": CSRFToken,
			"Message":    utils.ValidationErrorMsg(err.Error()),
		})
	}

	c.Response().Header().Set("HX-Redirect", "/admin/panel/users")
	return c.NoContent(http.StatusOK)
}
