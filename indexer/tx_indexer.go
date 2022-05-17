package indexer

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/blockchain"
	"github.com/tamas-soos/protocol-user-indexer/model"
	"github.com/tamas-soos/protocol-user-indexer/store"
)

func RunTxIndexers(store *store.Store, blockchain *blockchain.Client) {
	latestBlock, err := blockchain.BlockNumber(context.TODO())
	if err != nil {
		log.Fatal().Msgf("can't get latest block: %v", err)
	}

	networkID, err := blockchain.NetworkID(context.TODO())
	if err != nil {
		log.Fatal().Msgf("can't get network id: %v", err)
	}

	txIndexers, err := store.GetTxIndexers()
	if err != nil {
		log.Fatal().Msgf("can't get tx indexers: %v", err)
	}

	var wg sync.WaitGroup
	for _, txi := range txIndexers {
		txi := txi
		wg.Add(1)
		go func() {
			defer wg.Done()
			index(store, blockchain, txi, networkID, latestBlock)
		}()
	}

	wg.Wait()
}

func index(
	store *store.Store,
	blockchain *blockchain.Client,
	txi model.TxIndexer,
	networkID *big.Int,
	latestBlock uint64) {
	lastBlockIndexed := txi.LastBlockIndexed

	for lastBlockIndexed <= latestBlock-BATCH_SIZE {
		from, to := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE

		blocks, err := blockchain.BlocksByRange(from, to)
		if err != nil {
			log.Fatal().Msgf("can't get block: %v", err)
		}

		addresses, err := process(txi, blocks, networkID)
		if err != nil {
			log.Fatal().Msgf("can't process blocks: %v", err)
		}

		for _, address := range addresses {
			err := store.PutProtocolUser(txi.ID, address)
			if err != nil {
				log.Fatal().Msgf("can't store addresses: %v", err)
			}
		}

		err = store.UpdateLastBlockIndexedByID(txi.ID, lastBlockIndexed)
		if err != nil {
			log.Fatal().Msgf("can't update last block indexed: %v", err)
		}

		lastBlockIndexed = to

		log.Debug().
			Str("type", "tx").
			Int("protocol-indexer-id", txi.ID).
			Int("num-of-addresses", len(addresses)).
			Uint64("latest-block-indexed", lastBlockIndexed).
			Msg("indexing...")
	}

	log.Debug().
		Str("type", "tx").
		Int("protocol-indexer-id", txi.ID).
		Msg("indexer done")
}

func process(
	txi model.TxIndexer,
	blocks []*types.Block,
	networkID *big.Int) ([]string, error) {
	var addresses []string

	for _, block := range blocks {
		for _, tx := range block.Transactions() {
			// match condition
			if tx.To() != nil && txi.Spec.Condition.Tx.To == tx.To().String() {
				// select user
				if txi.Spec.User.Tx == "from" {
					msg, err := tx.AsMessage(types.LatestSignerForChainID(networkID), block.BaseFee())
					if err != nil {
						return nil, err
					}
					addresses = append(addresses, msg.From().Hex())
				}
			}
		}
	}

	return addresses, nil
}
