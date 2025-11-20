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

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	portStr := os.Getenv("port")
	if portStr == "" {
		log.Fatal("error: couldn't find the port in environment")
	}

	webCfg, err := web.NewWebConfig()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := webCfg.Server.DB.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	e := echo.New()

	// global set up
	e.Use(middleware.Logger())
	e.Validator = utils.NewCustomValidator()
	e.Renderer = newTemplate()
	e.Static("/static", "static")

	// main route (root)
	mainRoute := e.Group("")
	mainRoute.Use(webCfg.MiddlewareSession)
	mainRoute.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		CookiePath:     "/",
		TokenLength:    32,
		TokenLookup:    "header:X-CSRF-TOKEN,form:_csrf",
		ContextKey:     "csrf",
		CookieName:     "_csrf",
		CookieMaxAge:   86400,
		CookieHTTPOnly: true,
	}))

	mainRoute.Use(webCfg.MiddlewareAuthN)
	mainRoute.Use(webCfg.Server.MiddlewareAuthZ)
	mainRoute.Use(utils.MiddlewareUserRateLimiter)
	mainRoute.Use(utils.MiddlewareAPIRateLimiter)

	mainRoute.GET("/login", webCfg.GetLoginPage)
	mainRoute.POST("/login", webCfg.Login("/", utils.USER_ROLE_STUDENT))
	mainRoute.POST("/logout", webCfg.Logout("/login"))

	mainRoute.GET("/", webCfg.GetHomePage)

	mainRoute.GET("/students/:id/profile", webCfg.GetStudentProfile)
	mainRoute.GET("/students/:id/profile/update", webCfg.GetUpdateStudentPage)
	mainRoute.PUT("/students/:id/profile/update", webCfg.UpdateStudent)

	// admin route
	adminRoute := mainRoute.Group("/admin")
	adminRoute.GET("/login", webCfg.GetAdminLoginPage)
	adminRoute.POST("/login", webCfg.Login("/admin/panel", utils.USER_ROLE_ADMIN))
	adminRoute.POST("/logout", webCfg.Logout("/admin/login"))

	adminRoute.GET("/panel", webCfg.GetAdminPanelPage)

	adminRoute.GET("/panel/users", webCfg.GetUsersPage)
	adminRoute.POST("/panel/users/create", webCfg.CreateUser)

	adminRoute.GET("/panel/students", webCfg.GetStudentsPage)
	adminRoute.GET("/panel/students/create", webCfg.GetStudentSubmitPage)
	adminRoute.POST("/panel/students/create", webCfg.CreateStudent, webCfg.MiddlewareStudent)
	adminRoute.GET("/panel/students/:id/view", webCfg.GetStudentProfile)
	adminRoute.DELETE("/panel/students/:id/delete", webCfg.DeleteStudent)

	// SPAWN LIMITER CONTAINERS CLEANUP GOROUTINE
	utils.CleanupLimiterContainersWatcher()

	// SPAWN STALE USER SESSION CLEANER
	webCfg.Server.CleanStaleUserSessions()

	e.Logger.Fatal(e.Start(":" + portStr))
}
