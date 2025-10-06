package main

import (
	"html/template"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler/app"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

// helper to render html
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

	appCfg, err := app.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}

	defer appCfg.Server.DB.Close()

	e := echo.New()
	e.Use(middleware.Logger())

	e.Validator = utils.NewCustomValidator()
	e.Renderer = newTemplate()
	e.Static("/static", "static")

	e.GET("/", appCfg.GetHomePage)

	e.GET("/students", appCfg.GetStudentsPage)
	e.GET("/students/submission", appCfg.GetStudentSubmitPage)
	e.POST("/students/submission", appCfg.CreateStudent, appCfg.MiddlewareStudent)
	e.GET("/students/:id/profile", appCfg.GetStudentProfile)
	e.GET("/students/:id/profile/update", appCfg.GetUpdateStudentPage)
	e.PUT("/students/:id/profile/update", appCfg.UpdateStudent)
	e.DELETE("/students/:id/profile", appCfg.DeleteStudent)

	e.Logger.Fatal(e.Start(":" + portStr))

}
