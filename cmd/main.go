package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Custom Validator
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func main() {

	godotenv.Load(".env")

	portStr := os.Getenv("port")
	if portStr == "" {
		log.Fatal("error: couldn't find the port in environment")
	}

	server, err := handler.GetServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	defer server.DB.Close()

	e := echo.New()
	e.Use(middleware.Logger())

	// set custom validator to global scope
	// create custom oneof for validate tag
	v := validator.New()
	v.RegisterValidation("oneof_major", func(fl validator.FieldLevel) bool {
		majorStr := fl.Field().String()
		return slices.Contains(handler.MAJOR, majorStr)
	})

	v.RegisterValidation("oneof_room", func(fl validator.FieldLevel) bool {
		roomStr := fl.Field().String()
		return slices.Contains(handler.ROOM, roomStr)
	})

	e.Validator = &CustomValidator{validator: v}

	// essentially for this project, the CORS Config wouldnt be triggered.
	// because there is no Cross-Origin Resource Sharing.
	// the project use Same-origin to serve backend and frontend.

	// idk, it just nice to put here, i dont forget how to config it
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodDelete,
			http.MethodPut,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	e.Renderer = newTemplate()
	e.Static("/static", "static")

	e.GET("/", handler.HomeHandler)

	e.GET("/students", server.GetStudentsPage)

	e.GET("/students/submission", server.GetStudentSubmitPage)
	e.POST("/students/submission", server.CreateStudent, server.MiddlewareStudent)

	e.GET("/students/:id/profile", server.GetStudentProfile)
	e.DELETE("/students/:id/profile", server.DeleteStudent)
	e.GET("/students/:id/profile/update", server.GetUpdateStudentPage)
	e.PUT("/students/:id/profile/update", server.UpdateStudent)

	e.Logger.Fatal(e.Start(":" + portStr))

}
