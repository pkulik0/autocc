package main

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pkulik0/autocc/api/internal/auth"
	"github.com/pkulik0/autocc/api/internal/oauth"
	"github.com/pkulik0/autocc/api/internal/server"
	"github.com/pkulik0/autocc/api/internal/service"
	"github.com/pkulik0/autocc/api/internal/store"
	"github.com/pkulik0/autocc/api/internal/version"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Second
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	version.EnsureSet()
	log.Info().Str("version", version.Version).Str("build_time", version.BuildTime).Msg("AutoCC API")

	c, err := parseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config")
	}

	store, err := store.New(c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPass, c.PostgresDB)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create store")
	}
	oauth := oauth.New(c.GoogleCallbackURL)
	service := service.New(store, oauth)

	auth, err := auth.New(context.Background(), c.KeycloakURL, c.KeycloakRealm, c.KeycloakClientId, c.KeycloakClientSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create auth")
	}

	server := server.New(service, auth, c.GoogleRedirectURL)
	err = server.Start(c.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
