package store

import (
	"auth-as-a-service/app/memory/store/user"

	"github.com/jmoiron/sqlx"
)

// Registry holds every domain store. Add new stores here — server.go never changes.
type Registry struct {
	Users *user.Store
}

func New(db *sqlx.DB) *Registry {
	return &Registry{
		Users: user.NewStore(db),
	}
}
