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
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler/web"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/utils"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

// example
func main() {

	godotenv.Load(".env")

	portStr := os.Getenv("port")
	if portStr == "" {
		log.Fatal("error: couldn't find the port in environment")
	}

	webCfg, err := web.NewWebConfig()
	if err != nil {
		log.Fatal(err)
	}

	defer webCfg.Server.DB.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(50)))

	e.Validator = utils.NewCustomValidator()
	e.Renderer = newTemplate()
	e.Static("/static", "static")

	e.GET("/login", webCfg.GetLoginPage)

	e.GET("/", webCfg.GetHomePage)
	e.GET("/students", webCfg.GetStudentsPage)
	e.GET("/students/submission", webCfg.GetStudentSubmitPage)
	e.POST("/students/submission", webCfg.CreateStudent, webCfg.MiddlewareStudent)
	e.GET("/students/:id/profile", webCfg.GetStudentProfile)
	e.GET("/students/:id/profile/update", webCfg.GetUpdateStudentPage)
	e.PUT("/students/:id/profile/update", webCfg.UpdateStudent)
	e.DELETE("/students/:id/profile", webCfg.DeleteStudent)

	e.Logger.Fatal(e.Start(":" + portStr))

}
