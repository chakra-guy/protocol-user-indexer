package fetcher

import (
	"context"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/tamas-soos/protocol-user-indexer/internal/model"
	"google.golang.org/api/iterator"
)

type LogsBatchIterator struct {
	iterator *bigquery.RowIterator
}

func (f *Fetcher) QueryLogs(address string, from int) (*LogsBatchIterator, error) {
	query := f.client.Query(`
		SELECT topics as Topics, data as Data, block_number as BlockNumber
		FROM bigquery-public-data.crypto_ethereum.logs
		WHERE address = @contract_address AND block_number >= @from
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

	return &LogsBatchIterator{iterator: iterator}, nil
}

func (tbi *LogsBatchIterator) Next(ll *[]model.Log) (bool, error) {
	for {
		var l model.Log
		err := tbi.iterator.Next(&l)
		if err == iterator.Done {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		*ll = append(*ll, l)

		if tbi.iterator.PageInfo().Remaining() == 0 {
			return false, nil
		}
	}
}
