package main

import (
	"log"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/cmd/server"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/config"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/storage"
)

var (
	srvAddr                  = config.Env.ServerAddr
	PostgresConnStr          = config.Env.PostgresConnStr
	accessTokenSecret        = config.Env.AccessTokenSecret
	refreshTokenSecret       = config.Env.RefreshTokenSecret
	accessTokenExpiryInSecs  = config.Env.AccessTokenExpiryInSecs
	refreshTokenExpiryInSecs = config.Env.RefreshTokenExpiryInSecs
)

func main() {
	log.SetFlags(log.Ldate | log.Llongfile)

	db, err := storage.NewPostgresDB(PostgresConnStr)
	if err != nil {
		log.Fatal(err)
	}

	srv := server.NewServer(&server.ServerConfig{
		Addr: srvAddr,
		DB:   db,
		TokenManager: auth.NewTokenService(
			accessTokenSecret,
			refreshTokenSecret,
			accessTokenExpiryInSecs,
			refreshTokenExpiryInSecs,
		),
	},
	)
	srv.Run()
}
