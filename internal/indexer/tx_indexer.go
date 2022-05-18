package indexer

import (
	"context"
	"math/big"
	"sync"

	lo "github.com/samber/lo"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/internal/blockchain"
	"github.com/tamas-soos/protocol-user-indexer/internal/model"
	"github.com/tamas-soos/protocol-user-indexer/internal/store"
)

func RunTxIndexers(store *store.Store, blockchain *blockchain.Client) {
	networkID, err := blockchain.NetworkID(context.Background())
	if err != nil {
		log.Fatal().Msgf("can't get network id: %v", err)
	}

	latestBlock, err := blockchain.BlockNumber(context.Background())
	if err != nil {
		log.Fatal().Msgf("can't get latest block: %v", err)
	}

	txIndexers, err := store.GetTxIndexers()
	if err != nil {
		log.Fatal().Msgf("can't get tx indexers: %v", err)
	}

	partitions := lo.GroupBy(txIndexers, func(t model.TxIndexer) uint64 {
		return t.LastBlockIndexed
	})

	var wg sync.WaitGroup
	for id, partition := range partitions {
		id, partition := id, partition
		wg.Add(1)
		go func() {
			defer wg.Done()
			runPartion(store, blockchain, partition, networkID, id, latestBlock)
		}()
	}

	wg.Wait()
}

func runPartion(store *store.Store, blockchain *blockchain.Client, tt []model.TxIndexer, networkID *big.Int, startingBlock, latestBlock uint64) {
	log.Debug().Str("type", "tx").Uint64("partition-id", startingBlock).Int("num-of-indexers", len(tt)).Msg("running indexer partition...")

	lastBlockIndexed := startingBlock

	for lastBlockIndexed <= latestBlock-BATCH_SIZE {
		from, to := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE

		blocks, err := blockchain.BlocksByRange(from, to)
		if err != nil {
			log.Fatal().Msgf("can't get block: %v", err)
		}

		for _, txi := range tt {
			users, err := extractUsersFromTxs(txi, blocks, networkID)
			if err != nil {
				log.Fatal().Msgf("can't process blocks: %v", err)
			}

			err = store.PutProtocolUsers(txi.ID, users)
			if err != nil {
				log.Fatal().Msgf("can't store users: %v", err)
			}

			err = store.UpdateLastBlockIndexedByID(txi.ID, to)
			if err != nil {
				log.Fatal().Msgf("can't update last block indexed: %v", err)
			}

			lastBlockIndexed = to

			log.Debug().Str("type", "tx").Int("protocol-indexer-id", txi.ID).Int("num-of-users", len(users)).Uint64("latest-block-indexed", lastBlockIndexed).Msg("indexing...")
		}
	}

	log.Debug().Str("type", "tx").Uint64("partition-id", startingBlock).Int("num-of-indexers", len(tt)).Msg("indexer partition done")
}

func extractUsersFromTxs(txi model.TxIndexer, blocks []*types.Block, networkID *big.Int) ([]string, error) {
	var users []string

	for _, block := range blocks {
		for _, tx := range block.Transactions() {
			// match condition
			if tx.To() != nil && txi.Spec.Condition.Tx.To == tx.To().String() {
				// extract user
				if txi.Spec.User.Tx == "from" {
					msg, err := tx.AsMessage(types.LatestSignerForChainID(networkID), block.BaseFee())
					if err != nil {
						return nil, err
					}

					user := msg.From().Hex()
					users = append(users, user)
				}
			}
		}
	}

	return users, nil
}
