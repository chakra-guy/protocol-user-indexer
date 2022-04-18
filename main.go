package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	KEY = "EtoldX34DtXXzfknt1lNdbdGlPlZVU9T"
	URL = "https://eth-mainnet.alchemyapi.io/v2/" + KEY
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

	fmt.Println("LastProcessedBlock: ", block.Number())

	for _, tx := range block.Transactions() {
		if tx.To() == &p.ContractAddress {
			fmt.Println("uniswap tx: ", tx.Hash().Hex())

			chainID, err := client.NetworkID(context.TODO())
			if err != nil {
				fmt.Println("client.NetworkID err", err)
				return err
			}

			msg, err := tx.AsMessage(types.LatestSignerForChainID(chainID), block.BaseFee())
			if err != nil {
				fmt.Println("tx.AsMessage err", err)
				return err
			}

			fmt.Println("from: ", msg.From().Hex())

			fmt.Println()
		}
	}

	return nil
}

func main() {
	client, err := ethclient.Dial(URL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("we have a connection")

	uni := NewUniswapProtocol()

	start := time.Now()
	var wg sync.WaitGroup

	b := uni.LastProcessedBlock.Int64()
	for i := b; i <= b+100; i++ {
		wg.Add(1)
		go func(blockNumber int64) {
			defer wg.Done()

			block, err := client.BlockByNumber(context.TODO(), big.NewInt(blockNumber))
			if err != nil {
				log.Fatal("client.BlockByNumber err", err)
			}

			err = uni.ProcessNextBlock(client, block)
			if err != nil {
				log.Fatal("uni.ProcessNextBlock err", err)
			}
		}(i)
	}

	wg.Wait()

	fmt.Println("actaully done")

	elapsed := time.Since(start)
	log.Printf("Binomial took %s", elapsed)

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
