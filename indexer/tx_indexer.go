package indexer

import (
	"context"
	"encoding/json"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/model"
	"github.com/tamas-soos/protocol-user-indexer/store"
)

type TxIndexer struct {
	// deps
	store     *store.Store
	ethclient *ethclient.Client
	rpcclient *rpc.Client

	// FIXME should not be here, but fetch at the start
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

	// FIXME
	// latestBlock = txIndexers[0].LastBlockIndexed + BATCH_SIZE

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
}

func (indexer *TxIndexer) RunBatchProcessor(txi model.TxIndexer, latestBlock uint64) {
	lastBlockIndexed := txi.LastBlockIndexed

	// blocksCH := indexer.blockFetcherPool(lastBlockIndexed)

	// FIXME fix bug when indexer catches up and uses the wrong last indexed block number -> this can lead to skipping blocks
	for lastBlockIndexed <= latestBlock {
		// to := lastBlockIndexed + BATCH_SIZE
		// blocks := <-blocksCH

		from, to := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE

		blocks, err := indexer.fetchBlocksByRange(from, to)
		if err != nil {
			log.Fatal().Msgf("can't get block: %v", err)
		}

		addresses, err := indexer.processBlocks(txi, blocks)
		if err != nil {
			log.Fatal().Msgf("can't process blocks: %v", err)
		}

		err = indexer.storeResults(txi, addresses, to)
		if err != nil {
			log.Fatal().Msgf("can't store indexing results: %v", err)
		}

		lastBlockIndexed = to

		log.Debug().Str("type", "tx").Int("protocol-id", txi.ID).Int("num-of-addresses", len(addresses)).Uint64("latest-block-indexed", lastBlockIndexed).Send()
	}

	log.Debug().Str("type", "tx").Int("protocol-id", txi.ID).Msg("indexer caught up")
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

func (indexer *TxIndexer) fetchBlocksByRange(from, to uint64) ([]*types.Block, error) {
	var reqs []rpc.BatchElem
	rawblocks := make([]interface{}, BATCH_SIZE)
	index := 0

	for i := from; i <= to; i++ {
		reqs = append(reqs, rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{hexutil.EncodeBig(big.NewInt(int64(i))), true},
			Result: &rawblocks[index],
			// FIXME add error handling for each req
		})
		index++
	}

	err := indexer.rpcclient.BatchCall(reqs)
	if err != nil {
		return nil, err
	}

	var blocks []*types.Block
	for _, rawblock := range rawblocks {
		// FIXME just map things manually instead of doing this nonsense
		jsonblock, err := json.Marshal(rawblock)
		if err != nil {
			return nil, err
		}

		var head *types.Header
		err = json.Unmarshal(jsonblock, &head)
		if err != nil {
			return nil, err
		}

		var body struct {
			Transactions []*types.Transaction `json:"transactions"`
		}
		err = json.Unmarshal(jsonblock, &body)
		if err != nil {
			return nil, err
		}

		block := types.NewBlockWithHeader(head).WithBody(body.Transactions, nil)
		blocks = append(blocks, block)
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
