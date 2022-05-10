package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/config"
	"github.com/tamas-soos/wallet-explorer/db"
	"github.com/tamas-soos/wallet-explorer/server/handlers"
	"github.com/tamas-soos/wallet-explorer/store"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msg("starting server...")
	defer log.Info().Msg("ending server...")

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
