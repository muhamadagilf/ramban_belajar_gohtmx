// Package server
package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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

	return &Server{
		Queries: database.New(conn),
		DB:      conn,
	}, nil
}

func (server *Server) CleanStaleUserSessions() {
	log.Println("CLEANER RUNNNIG: Stale User Sessions")
	ticker := time.NewTicker(11 * time.Minute)
	ctx := context.Background()
	query := server.Queries

	go func() {
		for range ticker.C {
			log.Println("CLEANER CHECKPOINT: Stale User Sessions")
			expireTime := time.Now().Local().Add(-10 * time.Minute)
			sessions, err := query.GetSessionIDAll(ctx)
			if err != nil {
				log.Println(err)
			}

			for _, s := range sessions {
				if !s.LastActivity.Local().Add(-7 * time.Hour).After(expireTime) {
					if err := query.DeleteSessionByID(ctx, s.ID); err != nil {
						log.Println(err)
					}
				}
			}
		}
	}()
}
