package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

func (config *apiConfig) HandlerCreateUserAdmin(c echo.Context) error {
	var reqBody struct {
		Email    string `json:"email" validate:"email_constraints"`
		Password string `json:"password" validate:"password_constraints"`
	}

	err := utils.WithTX(c.Request().Context(), config.Server.DB, config.Server.Queries, func(qtx *database.Queries) error {
		if err := c.Bind(&reqBody); err != nil {
			return err
		}

		if err := c.Validate(&reqBody); err != nil {
			return err
		}

		passwordHashed, err := utils.HashPassword(reqBody.Password)
		if err != nil {
			return err
		}

		user, err := config.Server.Queries.CreateUser(c.Request().Context(), database.CreateUserParams{
			Email:        reqBody.Email,
			PasswordHash: passwordHashed,
		})
		if err != nil {
			return err
		}

		_, err = config.Server.Queries.CreateUserRoles(c.Request().Context(), database.CreateUserRolesParams{
			UserID: user.ID,
			Role:   utils.USER_ROLE_ADMIN,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Data{"message": err.Error()})
	}

	return c.JSON(http.StatusCreated, Data{"message": "User Created, Successfuly"})
}
