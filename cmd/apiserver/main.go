package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/muhamadagilf/rambanbelajar_gohtmx/handler/api"
)

func main() {

	godotenv.Load(".env")

	handlerFunc, err := api.NewApiConfig()
	if err != nil {
		log.Fatal(err)
	}

	defer handlerFunc.Server.DB.Close()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
		},
		AllowHeaders: []string{"*"},
	}))

	routerV1 := e.Group("/api/v1")
	routerV1.GET("/health", handlerFunc.HandlerHealth)
	routerV1.GET("/students", handlerFunc.HandlerGetStudents)

	e.Logger.Fatal(e.Start(":" + "8080"))

}
