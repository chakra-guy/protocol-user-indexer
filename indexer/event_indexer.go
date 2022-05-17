package indexer

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/blockchain"
	"github.com/tamas-soos/protocol-user-indexer/model"
	"github.com/tamas-soos/protocol-user-indexer/store"
)

func RunEventIndexer(store *store.Store, blockchain *blockchain.Client) {
	eventIndexers, err := store.GetEventIndexers()
	if err != nil {
		log.Fatal().Msgf("can't get event indexers: %v", err)
	}

	latestBlock, err := blockchain.BlockNumber(context.Background())
	if err != nil {
		log.Fatal().Msgf("can't get latest block: %v", err)
	}

	var wg sync.WaitGroup
	for _, ei := range eventIndexers {
		ei := ei
		wg.Add(1)
		go func() {
			defer wg.Done()
			batchIndexEvents(store, blockchain, ei, latestBlock)
		}()
	}

	wg.Wait()
}

func batchIndexEvents(store *store.Store, blockchain *blockchain.Client, ei model.EventIndexer, latestBlock uint64) {
	lastBlockIndexed := ei.LastBlockIndexed
	contractAddress := common.HexToAddress(ei.Spec.Condition.Contract.Address)
	contractABI, err := abi.JSON(strings.NewReader(ei.Spec.Condition.Contract.ABI))
	if err != nil {
		log.Warn().Int("protocol-id", ei.ID).Msgf("can't parse contract abi: %v", err)
		return
	}

	for lastBlockIndexed <= latestBlock-BATCH_SIZE {
		from, to := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE
		logs, err := blockchain.LogsByRange(from, to, contractAddress)
		if err != nil {
			log.Fatal().Msgf("can't get logs: %v", err)
		}

		users, err := extractUsersFromEvents(ei, logs, contractABI)
		if err != nil {
			log.Fatal().Msgf("can't process blocks: %v", err)
		}

		err = store.PutProtocolUsers(ei.ID, users)
		if err != nil {
			log.Fatal().Msgf("can't store users: %v", err)
		}

		err = store.UpdateLastBlockIndexedByID(ei.ID, lastBlockIndexed)
		if err != nil {
			log.Fatal().Msgf("can't update last block indexed: %v", err)
		}

		lastBlockIndexed = to

		log.Debug().Str("type", "event").Int("protocol-id", ei.ID).Int("num-of-users", len(users)).Uint64("latest-block-indexed", lastBlockIndexed).Msg("indexing...")
	}

	log.Debug().Str("type", "event").Int("protocol-id", ei.ID).Msg("indexer caught up")
}

func extractUsersFromEvents(ei model.EventIndexer, logs []types.Log, contractABI abi.ABI) ([]string, error) {
	var users []string

	for _, log := range logs {
		// match condition
		if log.Topics[0] == contractABI.Events[ei.Spec.Condition.Event.Name].ID && log.Address.String() == ei.Spec.Condition.Contract.Address {
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

func makeEvent(name string, contractABI abi.ABI, log types.Log) (map[string]interface{}, error) {
	event := make(map[string]interface{})

	err := contractABI.UnpackIntoMap(event, name, log.Data)
	if err != nil {
		return nil, err
	}

	var i = 1 // event args starts from index 1
	for _, input := range contractABI.Events[name].Inputs {
		if input.Indexed {
			if input.Type.String() == "address" {
				event[input.Name] = common.HexToAddress(log.Topics[i].Hex()).String()
			} else {
				// TODO handle more event arg types?
				event[input.Name] = log.Topics[i]
			}
			i += 1
		}
	}

	return event, nil
}
