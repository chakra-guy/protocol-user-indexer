package fetcher

import (
	"context"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/tamas-soos/protocol-user-indexer/internal/model"
	"google.golang.org/api/iterator"
)

type TransactionsBatchIterator struct {
	iterator *bigquery.RowIterator
}

func (f *Fetcher) QueryTransactions(address string, from int) (*TransactionsBatchIterator, error) {
	query := f.client.Query(`
		SELECT DISTINCT(from_address) as Sender, block_number as BlockNumber
		FROM bigquery-public-data.crypto_ethereum.transactions
		WHERE to_address = @contract_address AND block_number >= @from
		ORDER BY block_number ASC
	`)

	query.Parameters = []bigquery.QueryParameter{
		{Name: "contract_address", Value: strings.ToLower(address)},
		{Name: "from", Value: from},
	}

	query.Location = f.location

	iterator, err := query.Read(context.Background())
	if err != nil {
		return nil, err
	}

	return &TransactionsBatchIterator{iterator: iterator}, nil
}

func (tbi *TransactionsBatchIterator) Next(tt *[]model.Transaction) (bool, error) {
	for {
		var t model.Transaction
		err := tbi.iterator.Next(&t)
		if err == iterator.Done {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		*tt = append(*tt, t)

		if tbi.iterator.PageInfo().Remaining() == 0 {
			return false, nil
		}
	}
}
