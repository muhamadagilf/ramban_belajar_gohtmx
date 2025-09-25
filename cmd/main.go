package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

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

	e := echo.New()
	e.Use(middleware.Logger())

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

	e.GET("/student/profile/:id", server.GetStudentProfile)
	e.GET("/student/profile/:id/update", server.GetUpdateStudentPage)
	e.PUT("/student/profile/:id/update", server.UpdateStudent)
	e.DELETE("/student/profile/:id", server.DeleteStudent)

	e.Logger.Fatal(e.Start(":" + portStr))

}
