package main

import (
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/config"
	"github.com/tamas-soos/protocol-user-indexer/db"
	"github.com/tamas-soos/protocol-user-indexer/eth"
	"github.com/tamas-soos/protocol-user-indexer/indexer"
	"github.com/tamas-soos/protocol-user-indexer/store"
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
	ethclient := eth.New(&cfg.EthereumRPC)
	rpcclient, _ := rpc.Dial(cfg.EthereumRPC.URL + cfg.EthereumRPC.APIKey)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		indexer.NewTxIndexer(store, ethclient, rpcclient).Run()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		indexer.NewEventIndexer(store, ethclient, rpcclient).Run()
	}()

	wg.Wait()
}
