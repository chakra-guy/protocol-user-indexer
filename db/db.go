package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/config"
)

func New(cfg *config.Database) *sql.DB {
	log.Debug().Msg("connecting to database...")

	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		log.Fatal().Msgf("can't connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal().Msgf("can't connect to database: %v", err)
	}

	return db
}
