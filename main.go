package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/config"
	"github.com/tamas-soos/wallet-explorer/db"
	"github.com/tamas-soos/wallet-explorer/ethrpc"
	"github.com/tamas-soos/wallet-explorer/indexer/protocol"
	"github.com/tamas-soos/wallet-explorer/store"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal().Msgf("cannot load config variables: %v", err)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Info().Msg("starting worker...")
	defer log.Info().Msg("ending worker...")

	dbclient := db.New(&cfg.Database)
	store := store.New(dbclient)

	ethclient := ethrpc.New(&cfg.EthereumRPC)

	protocol.NewUniswap(store, ethclient).Index()
}
