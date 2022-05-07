package indexer

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/model"
	"github.com/tamas-soos/wallet-explorer/store"
)

var BATCH_SIZE uint64 = 10

type TxIndexer struct {
	// deps
	store     *store.Store
	ethclient *ethclient.Client
	rpcclient *rpc.Client

	// metadata (could be in config?)
	networkID *big.Int
}

func NewTxIndexer(store *store.Store, ethclient *ethclient.Client, rpcclient *rpc.Client) *TxIndexer {
	return &TxIndexer{
		store:     store,
		ethclient: ethclient,
		rpcclient: rpcclient,
	}
}

func (indexer *TxIndexer) Run() {
	latestBlock, err := indexer.ethclient.BlockNumber(context.TODO())
	if err != nil {
		log.Fatal().Msgf("can't get lastest block: %v", err)
	}

	indexer.networkID, err = indexer.ethclient.NetworkID(context.TODO())
	if err != nil {
		log.Fatal().Msgf("can't get network id: %v", err)
	}

	txIndexers, err := indexer.store.GetTxIndexers()
	if err != nil {
		log.Fatal().Msgf("can't get tx indexers: %v", err)
	}

	var wg sync.WaitGroup
	for _, txi := range txIndexers {
		txi := txi
		wg.Add(1)
		go func() {
			defer wg.Done()
			indexer.RunBatchProcessor(txi, latestBlock)
		}()
	}

	wg.Wait()

	_ = latestBlock
}

func (indexer *TxIndexer) RunBatchProcessor(txi model.TxIndexer, latestBlock uint64) {
	lastBlockIndexed := txi.LastBlockIndexed

	blocksCH := indexer.blockFetcherPool(lastBlockIndexed)

	for lastBlockIndexed <= latestBlock {
		to := lastBlockIndexed + BATCH_SIZE
		blocks := <-blocksCH

		// from, to := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE
		// blocks, err := indexer.fetchBlocksByRange(from, to)
		// if err != nil {
		// 	log.Fatal().Msgf("can't get block: %v", err)
		// }

		log.Debug().Msg("RunBatchProcessor -> processing blocks")
		addresses, err := indexer.processBlocks(txi, blocks)
		if err != nil {
			log.Fatal().Msgf("can't process blocks: %v", err)
		}

		log.Debug().Msgf("RunBatchProcessor -> storing results, addresses.len: %v, blocknumber: %v", len(addresses), to)
		err = indexer.storeResults(txi, addresses, to)
		if err != nil {
			log.Fatal().Msgf("can't store indexing results: %v", err)
		}

		lastBlockIndexed = to
	}
}

func (indexer *TxIndexer) blockFetcherPool(startBlock uint64) <-chan []*types.Block {
	POOL_SIZE := 5
	lastBlockIndexed := startBlock
	blocksCH := make(chan []*types.Block, 5)
	pool := make(chan struct{}, POOL_SIZE)

	for i := 0; i < POOL_SIZE; i++ {
		go func() {
			for {
				pool <- struct{}{}

				log.Debug().Msg("blockFetcherPool -> producing blocks")
				to, from := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE
				block, err := indexer.fetchBlocksByRange(to, from)
				if err != nil {
					log.Fatal().Msgf("can't get block: %v", err)
				}
				lastBlockIndexed = from
				blocksCH <- block

				<-pool
			}
		}()
	}

	return blocksCH
}

// func (indexer *TxIndexer) fetchBlocksByRange(from, to uint64) ([]*types.Block, error) {
// 	var payload []rpc.BatchElem
// 	blocks := make([]interface{}, BATCH_SIZE)
// 	index := 0

// 	for i := from; i <= to; i++ {
// 		payload = append(payload, rpc.BatchElem{
// 			Method: "eth_getBlockByNumber",
// 			Args:   []interface{}{hexutil.EncodeBig(big.NewInt(int64(i))), true},
// 			Result: &blocks[index],
// 		})
// 		index++
// 	}

// 	// var batchelements []map[string]interface{}
// 	// for i, p := range payload {
// 	// 	batchelements = append(batchelements, map[string]interface{}{
// 	// 		"jsonrpc": "2.0",
// 	// 		"id":      i,
// 	// 		"method":  p.Method,
// 	// 		"params":  p.Args,
// 	// 	})
// 	// }

// 	// stuff, _ := json.Marshal(batchelements)
// 	// fmt.Println(string(stuff))

// 	err := indexer.rpcclient.BatchCall(payload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return nil, nil
// }

func (indexer *TxIndexer) processBlocks(txi model.TxIndexer, blocks []*types.Block) ([]string, error) {
	var addresses []string

	// fmt.Printf("txi %+v\n", txi)
	// fmt.Printf("block %+v\n", blocks[0])

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
					msg, err := tx.AsMessage(types.LatestSignerForChainID(indexer.networkID), block.BaseFee())
					if err != nil {
						return nil, err
					}

					userAddress = msg.From().Hex()
					fmt.Println("userAddress", userAddress)
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
	for _, address := range addresses {
		err := indexer.store.PutProtocolUser(txi.ID, address)
		if err != nil {
			return err
		}
	}

	err := indexer.store.UpdateLastBlockIndexedByID(txi.ID, lastBlockIndexed)
	if err != nil {
		return err
	}

	return nil
}
