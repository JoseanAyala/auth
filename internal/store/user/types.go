package user

type User struct {
	ID           string `db:"id" json:"id"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"`
}
