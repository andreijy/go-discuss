package postgres

import (
	"fmt"

	godiscuss "github.com/andreijy/go-discuss"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func NewPostStore(db *sqlx.DB) *PostStore {
	return &PostStore{
		DB: db,
	}
}

type PostStore struct {
	*sqlx.DB
}

func (s *PostStore) Post(id uuid.UUID) (godiscuss.Post, error) {
	var p godiscuss.Post
	err := s.Get(&p, `SELECT * FROM posts WHERE id=$1`, id)
	if err != nil {
		return godiscuss.Post{}, fmt.Errorf("error getting post: %w", err)
	}
	return p, nil
}

func (s *PostStore) PostsByThread(threadID uuid.UUID) ([]godiscuss.Post, error) {
	var pp []godiscuss.Post
	var query = `
		SELECT 
			posts.*,
			COUNT(comments.*) AS comments_count
		FROM posts 
		LEFT JOIN comments ON comments.post_id = posts.id
		WHERE thread_id=$1 
		GROUP BY posts.id
		ORDER BY votes DESC`
	err := s.Select(&pp, query, threadID)
	if err != nil {
		return []godiscuss.Post{}, fmt.Errorf("error getting posts: %w", err)
	}
	return pp, nil
}

func (s *PostStore) Posts() ([]godiscuss.Post, error) {
	var pp []godiscuss.Post
	var query = `
		SELECT 
			posts.*,
			COUNT(comments.*) AS comments_count,
			threads.title AS thread_title
		FROM posts 
		LEFT JOIN comments ON comments.post_id = posts.id
		LEFT JOIN threads ON threads.id = posts.thread_id
		GROUP BY posts.id, threads.title
		ORDER BY votes DESC`
	err := s.Select(&pp, query)
	if err != nil {
		return []godiscuss.Post{}, fmt.Errorf("error getting posts: %w", err)
	}
	return pp, nil
}

func (s *PostStore) CreatePost(p *godiscuss.Post) error {
	err := s.Get(p, `INSERT INTO posts VALUES ($1, $2, $3, $4, $5) RETURNING *`,
		p.ID,
		p.ThreadID,
		p.Title,
		p.Content,
		p.Votes)
	if err != nil {
		return fmt.Errorf("error creating post: %w", err)
	}
	return nil
}

func (s *PostStore) UpdatePost(p *godiscuss.Post) error {
	err := s.Get(p, `UPDATE posts SET thread_id=$1, title=$2, content=$3, votes=$4 WHERE id=$5 RETURNING *`,
		p.ThreadID,
		p.Title,
		p.Content,
		p.Votes,
		p.ID)
	if err != nil {
		return fmt.Errorf("error updating post: %w", err)
	}
	return nil
}

func (s *PostStore) DeletePost(id uuid.UUID) error {
	_, err := s.Exec(`DELETE FROM posts WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}
	return nil
}
