package blockchain

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/internal/config"
)

type Client struct {
	*ethclient.Client
	rpc *rpc.Client
}

func New(cfg *config.EthereumRPC) *Client {
	log.Debug().Msg("connecting to ethereum rpc...")

	rpcclient, err := rpc.Dial(cfg.URL + cfg.APIKey)
	if err != nil {
		log.Fatal().Msgf("can't connect to ethereum rpc: %v", err)
	}

	ethclient := ethclient.NewClient(rpcclient)

	return &Client{
		Client: ethclient,
		rpc:    rpcclient,
	}
}

func (client *Client) BlocksByRange(from, to uint64) ([]*types.Block, error) {
	var reqs []rpc.BatchElem
	rawblocks := make([]interface{}, 10)
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

	err := client.rpc.BatchCall(reqs)
	if err != nil {
		return nil, err
	}

	// FIXME just map things manually instead of doing this 💩
	var blocks []*types.Block
	for _, rawblock := range rawblocks {
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

func (client *Client) LogsByRange(from, to uint64, address common.Address) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{address},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
