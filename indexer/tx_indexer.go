package indexer

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/model"
	"github.com/tamas-soos/wallet-explorer/store"
)

var BATCH_SIZE uint64 = 100

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
			indexer.RunBatchProcessor(txi, txi.LastBlockIndexed+(BATCH_SIZE*2))
		}()
	}

	wg.Wait()

	_ = latestBlock
}

func (indexer *TxIndexer) RunBatchProcessor(txi model.TxIndexer, latestBlock uint64) {
	lastBlockIndexed := txi.LastBlockIndexed

	for lastBlockIndexed <= latestBlock {
		to, from := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE
		blocks, err := indexer.fetchBlocksByRange(to, from)
		if err != nil {
			log.Fatal().Msgf("can't get block: %v", err)
		}

		addresses, err := indexer.processBlocks(txi, blocks)
		if err != nil {
			log.Fatal().Msgf("can't process blocks: %v", err)
		}

		err = indexer.storeResults(txi, addresses, from)
		if err != nil {
			log.Fatal().Msgf("can't store indexing results: %v", err)
		}

		lastBlockIndexed = from
	}
}

func (indexer *TxIndexer) fetchBlocksByRange(from, to uint64) ([]*types.Block, error) {
	fmt.Println("started fetching blocks")

	blocks := make([]*types.Block, BATCH_SIZE)
	var payload []rpc.BatchElem

	index := 0
	for i := from; i <= to; i++ {
		payload = append(payload, rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{hexutil.EncodeBig(big.NewInt(int64(i))), true},
			Result: &blocks[index],
		})
		index++
	}

	err := indexer.rpcclient.BatchCall(payload)
	if err != nil {
		return nil, err
	}

	fmt.Println("ended fetching blocks")

	return blocks, nil
}

func (indexer *TxIndexer) processBlocks(txi model.TxIndexer, blocks []*types.Block) ([]string, error) {
	fmt.Println("started processing blocks")

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
					msg, err := tx.AsMessage(types.LatestSignerForChainID(indexer.networkID), block.BaseFee())
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

	fmt.Println("ended processing blocks")

	return addresses, nil
}

func (indexer *TxIndexer) storeResults(txi model.TxIndexer, addresses []string, lastBlockIndexed uint64) error {
	fmt.Println("started storing results")

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

	fmt.Println("ended storing results")

	return nil
}
