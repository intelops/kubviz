package graph

import (
	"database/sql"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB *sql.DB
}

func NewResolver(db *sql.DB) *Resolver {
	return &Resolver{DB: db}
}
