package main

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/internal/blockchain"
	"github.com/tamas-soos/protocol-user-indexer/internal/config"
	"github.com/tamas-soos/protocol-user-indexer/internal/db"
	"github.com/tamas-soos/protocol-user-indexer/internal/indexer"
	"github.com/tamas-soos/protocol-user-indexer/internal/store"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msg("starting worker...")
	defer log.Info().Msg("ending worker...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Msgf("can't load config variables: %v", err)
	}

	db := db.New(&cfg.Database)
	store := store.New(db)
	blockchain := blockchain.New(&cfg.EthereumRPC)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		indexer.RunTxIndexers(store, blockchain)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		indexer.RunEventIndexer(store, blockchain)
	}()

	wg.Wait()
}
