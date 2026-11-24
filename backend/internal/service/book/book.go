package book

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	bookrepo "opd/internal/repository/book"
)

func (s *bookService) Upload(ctx context.Context, input UploadInput) (UploadResult, error) {
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return UploadResult{}, fmt.Errorf("generate object key: %w", err)
	}
	key := hex.EncodeToString(keyBytes)

	if err := s.storage.Put(ctx, s.booksBucket, key, input.Book.ContentType, input.Book.Body, input.Book.Size); err != nil {
		return UploadResult{}, fmt.Errorf("upload book: %w", err)
	}
	if err := s.storage.Put(ctx, s.coversBucket, key, input.Cover.ContentType, input.Cover.Body, input.Cover.Size); err != nil {
		s.cleanup(s.booksBucket, key)
		return UploadResult{}, fmt.Errorf("upload cover: %w", err)
	}
	id, err := s.repo.Create(ctx, bookrepo.CreateInput{
		Title: input.Title, Author: input.Author, Description: input.Description,
		AccessLevel: input.AccessLevel, Level: input.Level, Topic: input.Topic,
		BookKey: key, CoverKey: key, BookFilename: input.Book.Filename,
	})
	if err != nil {
		s.cleanup(s.booksBucket, key)
		s.cleanup(s.coversBucket, key)
		return UploadResult{}, fmt.Errorf("save book: %w", err)
	}
	return UploadResult{
		ID: id, Title: input.Title, Author: input.Author, Description: input.Description,
		AccessLevel: input.AccessLevel, Level: input.Level, Topic: input.Topic, BookKey: key, CoverKey: key,
	}, nil
}

func (s *bookService) cleanup(bucket, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.storage.Delete(ctx, bucket, key); err != nil {
		log.Printf("cleanup %s/%s: %v", bucket, key, err)
	}
}

func (s *bookService) List(ctx context.Context, input ListInput) (ListResult, error) {
	books, err := s.repo.List(ctx, bookrepo.ListInput{Limit: input.Limit + 1, Offset: input.Offset, Search: input.Search, Topic: input.Topic})
	if err != nil {
		return ListResult{}, fmt.Errorf("list books: %w", err)
	}
	hasMore := len(books) > input.Limit
	if hasMore {
		books = books[:input.Limit]
	}
	result := make([]Book, 0, len(books))
	for _, book := range books {
		result = append(result, Book{ID: book.ID, Title: book.Title, Author: book.Author, Description: book.Description, AccessLevel: book.AccessLevel, Level: book.Level, Topic: book.Topic})
	}
	return ListResult{Books: result, HasMore: hasMore}, nil
}

func (s *bookService) Download(ctx context.Context, id string) (DownloadResult, error) {
	book, err := s.repo.Get(ctx, id)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("get book: %w", err)
	}
	body, contentType, size, err := s.storage.Get(ctx, s.booksBucket, book.BookKey)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("download book: %w", err)
	}
	return DownloadResult{Body: body, ContentType: contentType, Size: size, Filename: book.BookFilename, Title: book.Title}, nil
}

func (s *bookService) Delete(ctx context.Context, id string) error {
	book, err := s.repo.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get book: %w", err)
	}
	if err := s.storage.Delete(ctx, s.booksBucket, book.BookKey); err != nil {
		return fmt.Errorf("delete book file: %w", err)
	}
	if err := s.storage.Delete(ctx, s.coversBucket, book.CoverKey); err != nil {
		return fmt.Errorf("delete book cover: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete book record: %w", err)
	}
	return nil
}
