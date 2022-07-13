package postgres

import (
	"fmt"

	godiscuss "github.com/andreijy/go-discuss"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{
		DB: db,
	}
}

type UserStore struct {
	*sqlx.DB
}

func (s *UserStore) User(id uuid.UUID) (godiscuss.User, error) {
	var u godiscuss.User
	err := s.Get(&u, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return godiscuss.User{}, fmt.Errorf("error getting user: %w", err)
	}
	return u, nil
}

func (s *UserStore) UserByUsername(username string) (godiscuss.User, error) {
	var u godiscuss.User
	err := s.Get(&u, `SELECT * FROM users WHERE username = $1`, username)
	if err != nil {
		return godiscuss.User{}, fmt.Errorf("error getting user by username: %w", err)
	}
	return u, nil
}

func (s *UserStore) Users() ([]godiscuss.User, error) {
	var uu []godiscuss.User
	err := s.Select(&uu, `SELECT * FROM users`)
	if err != nil {
		return []godiscuss.User{}, fmt.Errorf("error getting users: %w", err)
	}
	return uu, nil
}

func (s *UserStore) CreateUser(u *godiscuss.User) error {
	err := s.Get(u, `INSERT INTO users VALUES ($1, $2, $3) RETURNING *`,
		u.ID,
		u.Username,
		u.Password)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

func (s *UserStore) UpdateUser(u *godiscuss.User) error {
	err := s.Get(u, `UPDATE users SET user=$1, password=$2 WHERE id=$3 RETURNING *`,
		u.Username,
		u.Password,
		u.ID)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}

func (s *UserStore) DeleteUser(id uuid.UUID) error {
	_, err := s.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}
