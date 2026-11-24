package book

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type CreateInput struct {
	Title        string
	Author       string
	Description  string
	AccessLevel  string
	Level        string
	Topic        string
	BookKey      string
	BookFilename string
	CoverKey     string
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateInput) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `INSERT INTO books (title, author, description, access_level, level, topic, book_key, cover_key, book_filename) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id::text`, input.Title, input.Author, input.Description, input.AccessLevel, input.Level, input.Topic, input.BookKey, input.CoverKey, input.BookFilename).Scan(&id)
	return id, err
}

type ListInput struct {
	Search string
	Topic  string
	Limit  int
	Offset int
}

type Book struct {
	ID          string
	Title       string
	Author      string
	Description string
	AccessLevel string
	Level       string
	Topic       string
}

func (r *PostgresRepository) List(ctx context.Context, input ListInput) ([]Book, error) {
	rows, err := r.db.Query(ctx, `
  SELECT id::text, title, author, description, access_level, level, topic
  FROM books
  WHERE ($3 = '' OR strpos(lower(title), lower($3)) > 0 OR strpos(lower(author), lower($3)) > 0 OR strpos(lower(topic), lower($3)) > 0)
    AND ($4 = '' OR lower(topic) = lower($4))
  ORDER BY created_at DESC, id DESC
  LIMIT $1 OFFSET $2`, input.Limit, input.Offset, input.Search, input.Topic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	books := make([]Book, 0)
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Description, &book.AccessLevel, &book.Level, &book.Topic); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return books, nil
}

var ErrNotFound = errors.New("book not found")

type StoredBook struct {
	Title        string
	BookKey      string
	CoverKey     string
	BookFilename string
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (StoredBook, error) {
	var book StoredBook
	err := r.db.QueryRow(ctx, `SELECT title, book_key, cover_key, book_filename FROM books WHERE id = $1::uuid`, id).Scan(&book.Title, &book.BookKey, &book.CoverKey, &book.BookFilename)
	if errors.Is(err, pgx.ErrNoRows) {
		return StoredBook{}, ErrNotFound
	}
	return book, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM books WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
