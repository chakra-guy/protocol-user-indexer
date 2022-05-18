package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/internal/config"
	"github.com/tamas-soos/protocol-user-indexer/internal/db"
	"github.com/tamas-soos/protocol-user-indexer/internal/handlers"
	"github.com/tamas-soos/protocol-user-indexer/internal/store"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msg("starting api...")
	defer log.Info().Msg("ending api...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Msgf("can't load config variables: %v", err)
	}

	app := fiber.New()
	db := db.New(&cfg.Database)
	store := store.New(db)
	h := handlers.New(store)

	api := app.Group("/api")

	api.Get("/protocols", h.ListProtocols)
	api.Get("/:address/protocols", h.ListProtocolsByAddress)

	log.Fatal().Err(app.Listen(":3000"))
}
