package indexer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/internal/fetcher"
	"github.com/tamas-soos/protocol-user-indexer/internal/model"
	"github.com/tamas-soos/protocol-user-indexer/internal/store"
)

func RunEventIndexer(store *store.Store, fetcher *fetcher.Fetcher) {

	eventIndexers, err := store.GetEventIndexers()
	if err != nil {
		log.Fatal().Msgf("can't get event indexers: %v", err)
	}

	var wg sync.WaitGroup
	for _, ei := range eventIndexers {
		ei := ei
		wg.Add(1)
		go func() {
			defer wg.Done()

			log.Debug().Str("type", "event").Int("indexer-id", ei.ID).Int("starting-block", ei.LastBlockIndexed).Msg("running indexer...")
			start := time.Now()

			indexEvents(ei, store, fetcher)

			took := fmt.Sprintf("%.2f", time.Since(start).Minutes())
			log.Debug().Str("type", "event").Int("protocol-id", ei.ID).Str("took-min", took).Msg("indexer caught up")
		}()
	}

	wg.Wait()
}

func indexEvents(ei model.EventIndexer, store *store.Store, fetcher *fetcher.Fetcher) {

	contractABI, err := abi.JSON(strings.NewReader(ei.Spec.Condition.Contract.ABI))
	if err != nil {
		log.Warn().Int("protocol-id", ei.ID).Msgf("can't parse contract abi: %v", err)
		return
	}

	batches, err := fetcher.QueryLogs(ei.Spec.Condition.Contract.Address, ei.LastBlockIndexed)
	if err != nil {
		log.Fatal().Msgf("failed to assemble the query: %v", err)
	}

	for {
		start := time.Now()

		var ll []model.Log
		done, err := batches.Next(&ll)
		if err != nil {
			log.Fatal().Msgf("can't fetch logs: %v", err)
		}
		if done {
			break
		}

		users, err := extractUsersFromEvents(ei, ll, contractABI)
		if err != nil {
			log.Fatal().Msgf("can't process events: %v", err)
		}

		err = store.PutProtocolUsers(ei.ID, users)
		if err != nil {
			log.Fatal().Msgf("can't store users: %v", err)
		}

		lastBlockIndexed := ll[len(ll)-1].BlockNumber
		err = store.UpdateLastBlockIndexedByID(ei.ID, lastBlockIndexed)
		if err != nil {
			log.Fatal().Msgf("can't update last block indexed: %v", err)
		}

		took := fmt.Sprintf("%.2f", time.Since(start).Seconds())

		log.Debug().Str("type", "event").Int("protocol-id", ei.ID).Int("num-of-users", len(users)).Int("latest-block-indexed", lastBlockIndexed).Str("took-sec", took).Msg("indexing...")
	}
}

func extractUsersFromEvents(ei model.EventIndexer, logs []model.Log, contractABI abi.ABI) ([]string, error) {
	var users []string

	for _, log := range logs {
		// match condition
		if log.Topics[0] == contractABI.Events[ei.Spec.Condition.Event.Name].ID.String() {
			// extract user
			event, err := makeEvent(ei.Spec.Condition.Event.Name, contractABI, log)
			if err != nil {
				return nil, fmt.Errorf("can't make event from log: %v'", err)
			}

			user, ok := event[ei.Spec.User.Event.Arg].(string)
			if ok && user != "" {
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func makeEvent(name string, contractABI abi.ABI, log model.Log) (map[string]interface{}, error) {
	event := make(map[string]interface{})

	err := contractABI.UnpackIntoMap(event, name, common.FromHex(log.Data))
	if err != nil {
		return nil, err
	}

	var i = 1 // event args starts from index 1
	for _, input := range contractABI.Events[name].Inputs {
		if input.Indexed {
			if input.Type.String() == "address" {
				event[input.Name] = common.HexToAddress(log.Topics[i]).String()
			} else {
				// TODO handle more event arg types?
				event[input.Name] = log.Topics[i]
			}
			i += 1
		}
	}

	return event, nil
}
