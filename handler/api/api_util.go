package api

import "github.com/muhamadagilf/rambanbelajar_gohtmx/internal/server"

type apiConfig struct {
	Server *server.Server
}

func NewApiConfig() (*apiConfig, error) {
	server, err := server.GetServerConfig()
	if err != nil {
		return nil, err
	}

	return &apiConfig{
		Server: server,
	}, nil
}
