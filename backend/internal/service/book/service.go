package book

import (
	"context"
	"io"

	bookrepo "opd/internal/repository/book"
)

type BookService interface {
	Upload(ctx context.Context, input UploadInput) (UploadResult, error)
	List(ctx context.Context, input ListInput) (ListResult, error)
	Download(ctx context.Context, id string) (DownloadResult, error)
	Delete(ctx context.Context, id string) error
}

type bookRepository interface {
	Create(context.Context, bookrepo.CreateInput) (string, error)
	List(context.Context, bookrepo.ListInput) ([]bookrepo.Book, error)
	Get(context.Context, string) (bookrepo.StoredBook, error)
	Delete(context.Context, string) error
}

type fileStorage interface {
	Put(context.Context, string, string, string, io.ReadSeeker, int64) error
	Delete(context.Context, string, string) error
	Get(context.Context, string, string) (io.ReadCloser, string, int64, error)
}

type bookService struct {
	repo         bookRepository
	storage      fileStorage
	booksBucket  string
	coversBucket string
}

func NewBookService(repo bookRepository, storage fileStorage, booksBucket, coversBucket string) BookService {
	return &bookService{repo: repo, storage: storage, booksBucket: booksBucket, coversBucket: coversBucket}
}
