package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/tamas-soos/wallet-explorer/model"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (store Store) GetTxIndexers() ([]model.TxIndexer, error) {
	q := `
	SELECT id, last_block_indexed, spec FROM protocol_indexers
	WHERE type = 'tx'
	`
	rows, err := store.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tt := make([]model.TxIndexer, 0)

	for rows.Next() {
		var t model.TxIndexer
		var rawSpec []byte
		err := rows.Scan(&t.ID, &t.LastBlockIndexed, &rawSpec)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(rawSpec, &t.Spec)
		if err != nil {
			return nil, err
		}

		tt = append(tt, t)
	}

	return tt, nil
}

func (store Store) PutProtocolUser(protocolIndexerID int, address string) error {
	dbtx, err := store.db.BeginTx(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("can't insert protocol user into db: %v", err)
	}
	defer dbtx.Rollback()

	statement := `
	INSERT INTO users (address)
	VALUES ($1)
	ON CONFLICT (address) DO NOTHING
	`
	_, err = dbtx.Exec(statement, address)
	if err != nil {
		return fmt.Errorf("can't insert protocol user into db: %v", err)
	}

	statement2 := `
	INSERT INTO protocol_indexers_users (protocol_indexer_id, user_id)
	VALUES ($1, $2)
	ON CONFLICT (protocol_indexer_id, user_id) DO NOTHING
	`
	_, err = dbtx.Exec(statement2, protocolIndexerID, address)
	if err != nil {
		return fmt.Errorf("can't insert protocol user into db: %v", err)
	}

	if err = dbtx.Commit(); err != nil {
		return fmt.Errorf("can't insert protocol user into db: %v", err)
	}

	return nil
}

func (store Store) UpdateLastBlockIndexedByID(protocolIndexerID int, lastBlockIndexed uint64) error {
	statement := `
	UPDATE protocol_indexers
	SET last_block_indexed = $2
	WHERE id = $1
	`
	_, err := store.db.Exec(statement, protocolIndexerID, lastBlockIndexed)
	if err != nil {
		return err
	}

	return nil
}
