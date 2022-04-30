package store

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (store Store) SaveProtocolUser(protocolIndexerID int, address string) error {
	dbtx, err := store.db.BeginTx(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("cannot insert user address to db: %v", err)
	}
	defer dbtx.Rollback()

	statement := `
	INSERT INTO users (address)
	VALUES ($1)
	ON CONFLICT (address) DO NOTHING
	`
	_, err = dbtx.Exec(statement, address)
	if err != nil {
		return fmt.Errorf("cannot insert user address to db: %v", err)
	}

	statement2 := `
	INSERT INTO protocol_indexers_users (protocol_indexer_id, user_id)
	VALUES ($1, $2)
	ON CONFLICT (protocol_indexer_id, user_id) DO NOTHING
	`
	_, err = dbtx.Exec(statement2, protocolIndexerID, address)
	if err != nil {
		return fmt.Errorf("cannot insert user address to db: %v", err)
	}

	if err = dbtx.Commit(); err != nil {
		return fmt.Errorf("cannot insert user address to db: %v", err)
	}

	return nil
}
