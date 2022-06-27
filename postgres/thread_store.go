package postgres

import (
	"fmt"

	godiscuss "github.com/andreijy/go-discuss"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func NewThreadStore(db *sqlx.DB) *ThreadStore {
	return &ThreadStore{
		DB: db,
	}
}

type ThreadStore struct {
	*sqlx.DB
}

func (s *ThreadStore) Thread(id uuid.UUID) (godiscuss.Thread, error) {
	var t godiscuss.Thread
	err := s.Get(&t, `SELECT * FROM threads WHERE id = $1`, id)
	if err != nil {
		return godiscuss.Thread{}, fmt.Errorf("error getting thread: %w", err)
	}
	return t, nil
}

func (s *ThreadStore) Threads() ([]godiscuss.Thread, error) {
	var tt []godiscuss.Thread
	err := s.Select(&tt, `SELECT * FROM threads`)
	if err != nil {
		return []godiscuss.Thread{}, fmt.Errorf("error getting threads: %w", err)
	}
	return tt, nil
}

func (s *ThreadStore) CreateThread(t *godiscuss.Thread) error {
	err := s.Get(t, `INSERT INTO threads VALUES ($1, $2, $3) RETURNING *`,
		t.ID,
		t.Title,
		t.Description)
	if err != nil {
		return fmt.Errorf("error creating thread: %w", err)
	}
	return nil
}

func (s *ThreadStore) UpdateThread(t *godiscuss.Thread) error {
	err := s.Get(t, `UPDATE threads SET title=$1, description=$2 WHERE id=$3 RETURNING *`,
		t.Title,
		t.Description,
		t.ID)
	if err != nil {
		return fmt.Errorf("error updating thread: %w", err)
	}
	return nil
}

func (s *ThreadStore) DeleteThread(id uuid.UUID) error {
	_, err := s.Exec(`DELETE FROM threads WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting thread: %w", err)
	}
	return nil
}
