package protocol

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/store"
)

type UniswapIndexer struct {
	// hack
	id int

	// deps
	store     *store.Store
	ethclient *ethclient.Client

	// protocol specific config
	name               string
	deployBlock        *big.Int
	lastProcessedBlock *big.Int
	contractAddress    common.Address
}

func NewUniswap(store *store.Store, ethclient *ethclient.Client) *UniswapIndexer {
	return &UniswapIndexer{
		id:                 1,
		store:              store,
		ethclient:          ethclient,
		name:               "Uniswap",
		deployBlock:        big.NewInt(13804681),
		lastProcessedBlock: big.NewInt(13804682),
		contractAddress:    common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"),
	}
}

func (indexer *UniswapIndexer) Index() error {
	start := time.Now()
	var wg sync.WaitGroup

	b := indexer.lastProcessedBlock.Int64()
	for i := b; i <= b+100; i++ {
		wg.Add(1)
		go func(blockNumber int64) {
			defer wg.Done()

			block, err := indexer.ethclient.BlockByNumber(context.TODO(), big.NewInt(blockNumber))
			if err != nil {
				log.Fatal().Msgf("client.BlockByNumber: %v", err)
			}

			// log.Debug().Str("block", block.Number().String()).Send()

			for _, tx := range block.Transactions() {
				if tx.To() != nil && tx.To().Hex() == indexer.contractAddress.Hex() {
					// log.Debug().Str("uniswap tx", tx.Hash().Hex()).Send()

					chainID, err := indexer.ethclient.NetworkID(context.TODO())
					if err != nil {
						log.Fatal().Msgf("indexer.ethclient.NetworkID: %v", err)
					}

					msg, err := tx.AsMessage(types.LatestSignerForChainID(chainID), block.BaseFee())
					if err != nil {
						log.Fatal().Msgf(" tx.AsMessage: %v", err)
					}

					userAddress := msg.From().Hex()
					err = indexer.store.InsertUserAddress(indexer.id, userAddress)
					if err != nil {
						log.Fatal().Msgf("indexer.store.InsertUserAddress: %v", err)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)
	log.Debug().Msgf("100 block processing took: %s\n", elapsed)

	return nil
}
