package indexer

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/internal/fetcher"
	"github.com/tamas-soos/protocol-user-indexer/internal/model"
	"github.com/tamas-soos/protocol-user-indexer/internal/store"
)

func RunTxIndexers(store *store.Store, fetcher *fetcher.Fetcher) {

	txIndexers, err := store.GetTxIndexers()
	if err != nil {
		log.Fatal().Msgf("can't get tx indexers: %v", err)
	}

	var wg sync.WaitGroup
	for _, ti := range txIndexers {
		ti := ti
		wg.Add(1)
		go func() {
			defer wg.Done()

			start := time.Now()
			log.Debug().Str("type", "tx").Int("indexer-id", ti.ID).Int("starting-block", ti.LastBlockIndexed).Msg("running indexer...")

			indexTxs(ti, store, fetcher)

			took := fmt.Sprintf("%.2f", time.Since(start).Minutes())
			log.Debug().Str("type", "tx").Int("indexer-id", ti.ID).Str("took-min", took).Msg("indexer caught up")
		}()
	}

	wg.Wait()
}

func indexTxs(ti model.TxIndexer, store *store.Store, fetcher *fetcher.Fetcher) {

	batches, err := fetcher.QueryTransactions(ti.Spec.Condition.Tx.To, ti.LastBlockIndexed)
	if err != nil {
		log.Fatal().Msgf("failed to assemble the query: %v", err)
	}

	for {
		start := time.Now()

		var tt []model.Transaction
		done, err := batches.Next(&tt)
		if err != nil {
			log.Fatal().Msgf("can't fetch transactions: %v", err)
		}
		if done {
			break
		}

		var users []string
		for _, t := range tt {
			users = append(users, t.Sender)
		}

		err = store.PutProtocolUsers(ti.ID, users)
		if err != nil {
			log.Fatal().Msgf("can't store users: %v", err)
		}

		lastBlockIndexed := tt[len(tt)-1].BlockNumber
		err = store.UpdateLastBlockIndexedByID(ti.ID, lastBlockIndexed)
		if err != nil {
			log.Fatal().Msgf("can't update last block indexed: %v", err)
		}

		took := fmt.Sprintf("%.2f", time.Since(start).Seconds())

		log.Debug().Str("type", "tx").Int("indexer-id", ti.ID).Int("num-of-users", len(users)).Int("latest-block-indexed", lastBlockIndexed).Str("took-sec", took).Msg("indexing...")
	}
}
