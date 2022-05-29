package fetcher

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/rs/zerolog/log"
	"github.com/tamas-soos/protocol-user-indexer/internal/config"
	"google.golang.org/api/option"
)

type Fetcher struct {
	client   *bigquery.Client
	dataset  string
	location string
}

func New(cfg *config.GCP) *Fetcher {
	client, err := bigquery.NewClient(context.Background(), cfg.ProjectID, option.WithCredentialsJSON([]byte(cfg.ServiceAccount)))
	if err != nil {
		log.Fatal().Msgf("can't create bigquery client: %v", err)
	}

	return &Fetcher{
		client:   client,
		dataset:  cfg.BigQuery.Dataset,
		location: cfg.BigQuery.Location,
	}
}
