package eth

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/config"
)

func New(cfg *config.EthereumRPC) *ethclient.Client {
	log.Debug().Msg("connecting to ethereum rpc...")

	client, err := ethclient.Dial(cfg.URL + cfg.APIKey)
	if err != nil {
		log.Fatal().Msgf("can't connect to ethereum rpc: %v", err)
	}

	return client
}
