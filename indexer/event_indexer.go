package indexer

import (
	"context"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/model"
	"github.com/tamas-soos/protocol-user-indexer/store"
)

type EventIndexer struct {
	// deps
	store     *store.Store
	ethclient *ethclient.Client
	rpcclient *rpc.Client
}

func NewEventIndexer(store *store.Store, ethclient *ethclient.Client, rpcclient *rpc.Client) *EventIndexer {
	return &EventIndexer{
		store:     store,
		ethclient: ethclient,
		rpcclient: rpcclient,
	}
}

func (indexer *EventIndexer) Run() {
	latestBlock, err := indexer.ethclient.BlockNumber(context.TODO())
	if err != nil {
		log.Fatal().Msgf("can't get latest block: %v", err)
	}

	eventIndexers, err := indexer.store.GetEventIndexers()
	if err != nil {
		log.Fatal().Msgf("can't get event indexers: %v", err)
	}

	// FIXME
	// latestBlock = eventIndexers[0].LastBlockIndexed + (BATCH_SIZE)

	var wg sync.WaitGroup
	for _, ei := range eventIndexers {
		ei := ei
		wg.Add(1)
		go func() {
			defer wg.Done()
			indexer.RunBatchProcessor(ei, latestBlock)
		}()
	}

	wg.Wait()
}

func (indexer *EventIndexer) RunBatchProcessor(ei model.EventIndexer, latestBlock uint64) {
	contractABI, err := abi.JSON(strings.NewReader(ei.Spec.Condition.Contract.ABI))
	if err != nil {
		log.Warn().Int("protocol-id", ei.ID).Msgf("can't run indexer: can't parse contract abi: %v", err)
		return
	}

	lastBlockIndexed := ei.LastBlockIndexed

	// FIXME fix bug when indexer catches up and uses the wrong last indexed block number -> this can lead to skipping blocks
	for lastBlockIndexed <= latestBlock {
		from, to := lastBlockIndexed+1, lastBlockIndexed+BATCH_SIZE

		logs, err := indexer.fetchLogsByRange(ei, from, to)
		if err != nil {
			log.Fatal().Msgf("can't get logs: %v", err)
		}

		addresses, err := indexer.processLogs(ei, contractABI, logs)
		if err != nil {
			log.Fatal().Msgf("can't process logs: %v", err)
		}

		err = indexer.storeResults(ei, addresses, to)
		if err != nil {
			log.Fatal().Msgf("can't store indexing results: %v", err)
		}

		lastBlockIndexed = to

		log.Debug().Str("type", "event").Int("protocol-id", ei.ID).Int("num-of-addresses", len(addresses)).Uint64("latest-block-indexed", lastBlockIndexed).Send()
	}

	log.Debug().Str("type", "event").Int("protocol-id", ei.ID).Msg("indexer caught up")
}

func (indexer *EventIndexer) fetchLogsByRange(ei model.EventIndexer, from, to uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{common.HexToAddress(ei.Spec.Condition.Contract.Address)},
	}

	logs, err := indexer.ethclient.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (indexer *EventIndexer) processLogs(ei model.EventIndexer, contractABI abi.ABI, logs []types.Log) ([]string, error) {
	var addresses []string

	for _, log := range logs {
		if log.Topics[0] == contractABI.Events[ei.Spec.Condition.Event.Name].ID && log.Address.String() == ei.Spec.Condition.Contract.Address {
			event := make(map[string]interface{})

			// fmt.Printf("event name: %s\n", ei.Spec.Condition.Event.Name)
			// fmt.Printf("log: %+v\n", log)
			// fmt.Printf("log data: %+v\n", log.Data)

			err := contractABI.UnpackIntoMap(event, ei.Spec.Condition.Event.Name, log.Data)
			if err != nil {
				return nil, err
			}

			var i = 1 // topic_name_0 is the event signature so we start from 1
			for _, input := range contractABI.Events[ei.Spec.Condition.Event.Name].Inputs {
				if input.Indexed {
					if input.Type.String() == "address" {
						event[input.Name] = common.HexToAddress(log.Topics[i].Hex()).String()
					} else {
						// TODO handle more event ar types
						event[input.Name] = log.Topics[i]
					}
					i += 1
				}
			}

			address, ok := event[ei.Spec.User.Event.Arg].(string)
			if ok && address != "" {
				addresses = append(addresses, address)
			}
		}
	}

	return addresses, nil
}

func (indexer *EventIndexer) storeResults(ei model.EventIndexer, addresses []string, lastBlockIndexed uint64) error {
	for _, address := range addresses {
		err := indexer.store.PutProtocolUser(ei.ID, address)
		if err != nil {
			return err
		}
	}

	err := indexer.store.UpdateLastBlockIndexedByID(ei.ID, lastBlockIndexed)
	if err != nil {
		return err
	}

	return nil
}
