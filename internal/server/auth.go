package server

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	UserID uuid.UUID
	Roles  []string
}

var Permissions = map[string][]string{
	"admin": {
		"adminPanelPages:view",
		"users:*",
		"teachers:*",
		"students:*",
		"studentCreatePage:view",
	},
	"teacher": {
		"homePage:view",
		"coursePage:view",
		"teachers:view",
		"courses:*",
	},
	"student": {
		"homePage:view",
		"students:view",
	},
}

func matchPermission(permission, resource, action string) bool {
	if permission == "*" {
		return true
	}

	parts := strings.SplitN(permission, ":", 2)
	if len(parts) != 2 {
		return false
	}

	r, a := parts[0], parts[1]

	if r != resource {
		return false
	}

	return a == action || a == "*"
}

func (s *Server) Can(claims *Claims, resource, action string) (bool, string) {
	for _, role := range claims.Roles {
		permissions := Permissions[role]
		for _, perm := range permissions {
			if matchPermission(perm, resource, action) {
				return true, role
			}
		}
	}

	return false, ""
}

func (s *Server) LoadUserRoles(context context.Context, userID uuid.UUID) ([]string, error) {
	userRoles, err := s.Queries.GetUserRolesByUserID(context, userID)
	if err != nil {
		return nil, err
	}

	roles := []string{}
	for _, role := range userRoles {
		roles = append(roles, role.Role)
	}

	return roles, nil
}

func (s *Server) MiddlewareAuthZ(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		context := c.Request().Context()
		userID, ok := c.Get("user_id").(uuid.UUID)
		if !ok {
			return next(c)
		}

		roles, err := s.LoadUserRoles(context, userID)
		if err != nil {
			log.Println(err)
			return c.String(
				http.StatusInternalServerError,
				"Internal Server Error. Contact support with this code: ERR035001",
			)
		}

		c.Set("claims", &Claims{
			UserID: userID,
			Roles:  roles,
		})

		return next(c)
	}
}
