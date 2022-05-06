package indexer

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/model"
	"github.com/tamas-soos/wallet-explorer/store"
)

var BATCH_SIZE uint64 = 10

type TxIndexer struct {
	store     *store.Store
	ethclient *ethclient.Client
	chainID   *big.Int
}

func NewTxIndexer(store *store.Store, ethclient *ethclient.Client) *TxIndexer {
	return &TxIndexer{
		store:     store,
		ethclient: ethclient,
	}
}

func (indexer *TxIndexer) Run() {
	latestBlock, err := indexer.ethclient.BlockNumber(context.TODO())
	if err != nil {
		log.Fatal().Msgf("cannot get lastest block: %v", err)
	}

	indexer.chainID, err = indexer.ethclient.NetworkID(context.TODO())
	if err != nil {
		log.Fatal().Msgf("cannot get network id: %v", err)
	}

	txIndexers, err := indexer.store.GetTxIndexers()
	if err != nil {
		log.Fatal().Msgf("cannot get tx indexers: %v", err)
	}

	var wg sync.WaitGroup
	for _, txi := range txIndexers {
		txi := txi
		wg.Add(1)
		go func() {
			defer wg.Done()
			indexer.RunBatchProcessor(txi, txi.LastBlockIndexed+(BATCH_SIZE*2))
		}()
	}

	wg.Wait()

	_ = latestBlock
}

func (indexer *TxIndexer) RunBatchProcessor(txi model.TxIndexer, latestBlock uint64) {
	lastBlockIndexed := txi.LastBlockIndexed

	for lastBlockIndexed < latestBlock {
		blocks, err := indexer.fetchBlocksByRange(lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE)
		if err != nil {
			log.Fatal().Msgf("cannot get block: %v", err)
		}

		addresses, err := indexer.processBlocks(txi, blocks)
		if err != nil {
			log.Fatal().Msgf("cannot process blocks: %v", err)
		}

		err = indexer.storeResults(txi, addresses, lastBlockIndexed+BATCH_SIZE)
		if err != nil {
			log.Fatal().Msgf("cannot store indexing results: %v", err)
		}

		lastBlockIndexed += BATCH_SIZE
	}
}

// func (indexer *TxIndexer) fetchBlocksByRange(from, to uint64) ([]*types.Block, error) {
// 	var blocks []*types.Block

// 	results := make(chan *types.Block)
// 	failure := make(chan struct{})

// 	for i := from; i <= to; i++ {
// 		i := i
// 		go func() {
// 			block, err := indexer.ethclient.BlockByNumber(context.TODO(), big.NewInt(int64(i)))
// 			if err != nil {
// 				failure <- struct{}{}
// 			}
// 			results <- block
// 			fmt.Println("fetchBlocksByRange -> produce", block.Hash())
// 		}()
// 	}

// 	// TODO failure
// 	// TODO timeout

// 	for i := from; i <= to; i++ {
// 		block := <-results
// 		fmt.Println("fetchBlocksByRange -> consume", block.Hash())
// 		blocks = append(blocks, block)
// 	}

// 	return blocks, nil
// }

func (indexer *TxIndexer) fetchBlocksByRange(from, to uint64) ([]*types.Block, error) {
	var blocks []*types.Block

	for i := from; i <= to; i++ {
		block, err := indexer.ethclient.BlockByNumber(context.TODO(), big.NewInt(int64(i)))
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)

		fmt.Println("block:", block.Hash())
	}

	return blocks, nil
}

func (indexer *TxIndexer) processBlocks(txi model.TxIndexer, blocks []*types.Block) ([]string, error) {
	var addresses []string

	for _, block := range blocks {
		for _, tx := range block.Transactions() {
			if tx.To() == nil {
				continue
			}

			// check conditions
			if txi.Spec.Condition.Tx.To == tx.To().String() {
				var userAddress string
				// check user
				if txi.Spec.User.Tx == "from" {
					msg, err := tx.AsMessage(types.LatestSignerForChainID(indexer.chainID), block.BaseFee())
					if err != nil {
						return nil, err
					}

					userAddress = msg.From().Hex()
					if userAddress != "" {
						addresses = append(addresses, userAddress)
					}
				}
			}
		}
	}

	return addresses, nil
}

func (indexer *TxIndexer) storeResults(txi model.TxIndexer, addresses []string, lastBlockIndexed uint64) error {
	fmt.Println("started storing results")

	for _, address := range addresses {
		err := indexer.store.SaveProtocolUser(txi.ID, address)
		if err != nil {
			return err
		}
	}

	err := indexer.store.UpdateLastBlockIndexedByID(txi.ID, lastBlockIndexed)
	if err != nil {
		return err
	}

	fmt.Println("ended storing results")

	return nil
}
