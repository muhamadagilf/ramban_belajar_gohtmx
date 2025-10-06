package server

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/muhamadagilf/rambanbelajar_gohtmx/internal/database"
)

type Server struct {
	Queries *database.Queries
	DB      *sql.DB
}

func GetServerConfig() (*Server, error) {

	dbURL := os.Getenv("db_url")
	if dbURL == "" {
		return nil, fmt.Errorf("error: cannot find db_url in the environment")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	return &Server{Queries: database.New(conn), DB: conn}, nil

}
