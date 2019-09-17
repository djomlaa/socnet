package service

import (
	"github.com/hako/branca"
	"database/sql"
)


// Service contains the core logic
// Can be used to back Rest, GraphQL or RPC API
type Service struct {

	db *sql.DB
	codec *branca.Branca

}

// New Service implementation
func New(db *sql.DB, codec *branca.Branca) *Service {
	return &Service{db: db, codec: codec}
}