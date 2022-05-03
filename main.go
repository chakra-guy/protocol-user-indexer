package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/config"
	"github.com/tamas-soos/wallet-explorer/db"
	"github.com/tamas-soos/wallet-explorer/ethrpc"
	"github.com/tamas-soos/wallet-explorer/indexer"
	"github.com/tamas-soos/wallet-explorer/store"
	"github.com/tamas-soos/wallet-explorer/types"
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

	eventSpec := types.EventIndexerSpec{
		Condition: struct{ Event struct{ Name string } }{
			Event: struct{ Name string }{
				Name: "Approval",
			},
		},
		User: struct{ Event struct{ Arg string } }{
			Event: struct{ Arg string }{
				Arg: "owner",
			},
		},
	}

	indexer.NewTxIndexer(store, ethclient).Index()
	indexer.NewEventIndexer(store, ethclient).Index(eventSpec)
}
