package main

import (
	"context"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/wallet-explorer/config"
	"github.com/tamas-soos/wallet-explorer/db"
	"github.com/tamas-soos/wallet-explorer/ethereum_rpc"
)

type Protocol struct {
	Name               string
	ContractAddress    common.Address
	DeployBlock        *big.Int
	LastProcessedBlock *big.Int
	Users              []common.Address
}

func NewUniswapProtocol() *Protocol {
	return &Protocol{
		Name:               "Uniswap",
		ContractAddress:    common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"),
		DeployBlock:        big.NewInt(13804681),
		LastProcessedBlock: big.NewInt(13804682),
	}
}

func (p *Protocol) ProcessNextBlock(client *ethclient.Client, block *types.Block) error {
	log.Debug().Str("block", block.Number().String()).Send()

	for _, tx := range block.Transactions() {
		if tx.To() != nil && tx.To().Hex() == p.ContractAddress.Hex() {
			log.Debug().Str("uniswap tx", tx.Hash().Hex()).Send()

			chainID, err := client.NetworkID(context.TODO())
			if err != nil {
				return err
			}

			msg, err := tx.AsMessage(types.LatestSignerForChainID(chainID), block.BaseFee())
			if err != nil {
				return err
			}

			log.Debug().Str("from", msg.From().Hex()).Send()

		}
	}

	return nil
}

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal().Msgf("cannot load config variables: %v", err)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Info().Msg("starting worker...")
	defer log.Info().Msg("ending worker...")

	dbclient := db.New(&cfg.Database)
	ethclient := ethereum_rpc.New(&cfg.EthereumRPC)

	_ = dbclient
	_ = ethclient

	uni := NewUniswapProtocol()

	start := time.Now()
	var wg sync.WaitGroup

	b := uni.LastProcessedBlock.Int64()
	for i := b; i <= b+100; i++ {
		wg.Add(1)
		go func(blockNumber int64) {
			defer wg.Done()

			block, err := ethclient.BlockByNumber(context.TODO(), big.NewInt(blockNumber))
			if err != nil {
				log.Fatal().Msgf("client.BlockByNumber: %v", err)
			}

			err = uni.ProcessNextBlock(ethclient, block)
			if err != nil {
				log.Fatal().Msgf("uni.ProcessNextBlock: %v", err)
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)
	log.Debug().Msgf("100 block processing took: %s\n", elapsed)
}

// number := big.NewInt(14192475)
// block, err := client.BlockByNumber(context.Background(), number)
// if err != nil {
// 	log.Fatal(err)
// }

// for _, tx := range block.Transactions() {
// 	// https://etherscan.io/tx/0x7b989674ab01060a9ae6842291f8ccbd781e629a28deef7461f8afd512a63ed7
// 	if tx.Hash().Hex() == "0x7b989674ab01060a9ae6842291f8ccbd781e629a28deef7461f8afd512a63ed7" {
// 		fmt.Println("found it!")

// 		fmt.Println(tx.Hash().Hex())
// 		fmt.Println(tx.To().Hex())

// 		// msg, _ := tx.AsMessage(types.NewEIP2930Signer(tx.ChainId()), big.NewInt(1))
// 		// fmt.Println(msg.From().Hex())
// 		signer := types.NewEIP155Signer(tx.ChainId())
// 		sender, err := signer.Sender(tx)
// 		if err != nil {
// 			fmt.Printf("sender: %v", sender.Hex())
// 		}
// 	}
// }
