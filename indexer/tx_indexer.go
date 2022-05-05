package indexer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tamas-soos/wallet-explorer/store"
)

type TxIndexer struct {
	store     *store.Store
	ethclient *ethclient.Client
}

func NewTxIndexer(store *store.Store, ethclient *ethclient.Client) *TxIndexer {
	return &TxIndexer{
		store:     store,
		ethclient: ethclient,
	}
}

//
// currentBlock := ethclient.GetLatestBlockNumber()
//
// go indexer.NewTxIndexer(...deps).Run(currentBlock)		# RUN
//    indexers := indexer.store.GetTxIndexers()
//    for indexers
//        if lastBlockIndexed < currentBlock
//            blocks := ethclient.FetchBlocks()
//            for blocks
//                user = process(block, indexer)			# EXTRACT
//                if user
//                    store(user)
//             store(lastBlockIndexed)
//

func (indexer *TxIndexer) Run() error {
	txIndexer, err := indexer.store.GetTxIndexers()
	if err != nil {
		return err
	}

	fmt.Printf("txIndexer %+v\n\n", txIndexer)

	for _, txIndexer := range txIndexer {
		number := big.NewInt(int64(txIndexer.LastIndexedBlock))
		block, err := indexer.ethclient.BlockByNumber(context.TODO(), number)
		if err != nil {
			return err
		}

		for _, tx := range block.Transactions() {
			if tx.To() == nil {
				continue
			}

			// check conditions
			if txIndexer.Spec.Condition.Tx.To == tx.To().String() {
				var userAddress string

				// check user
				if txIndexer.Spec.User.Tx == "from" {
					chainID, err := indexer.ethclient.NetworkID(context.TODO())
					if err != nil {
						return err
					}

					msg, err := tx.AsMessage(types.LatestSignerForChainID(chainID), block.BaseFee())
					if err != nil {
						return err
					}

					userAddress = msg.From().Hex()
				}

				fmt.Println("user:", userAddress)

				// save
				if userAddress != "" {
					err = indexer.store.SaveProtocolUser(txIndexer.ID, userAddress)
					if err != nil {
						return err
					}
				}
			}

			// TODO update last block indexed
		}
	}

	return nil
}
