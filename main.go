package main

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/config"
	"github.com/tamas-soos/wallet-explorer/db"
	"github.com/tamas-soos/wallet-explorer/ethrpc"
	"github.com/tamas-soos/wallet-explorer/indexer"
	"github.com/tamas-soos/wallet-explorer/store"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal().Msgf("can't load config variables: %v", err)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Info().Msg("starting worker...")
	defer log.Info().Msg("ending worker...")

	dbclient := db.New(&cfg.Database)
	store := store.New(dbclient)
	ethclient := ethrpc.New(&cfg.EthereumRPC)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		indexer.NewTxIndexer(store, ethclient).Run()
	}()

	wg.Wait()
}
