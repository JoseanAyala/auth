package user

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, email, passwordHash string) (User, error) {
	var u User
	err := s.db.GetContext(ctx, &u,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email", email, passwordHash)
	return u, err
}

func (s *Store) GetByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.db.GetContext(ctx, &u, "SELECT id, email, password_hash FROM users WHERE email = $1", email)
	return u, err
}
