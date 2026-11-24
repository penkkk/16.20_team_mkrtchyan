package book

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	bookrepo "opd/internal/repository/book"
)

const testBooksBucket = "books"
const testCoversBucket = "covers"
const testPDFType = "application/pdf"

type fakeStorage struct {
	puts             []string
	deletes          []string
	failBucket       string
	deleteFailBucket string
	keys             []string
}

func (s *fakeStorage) Put(_ context.Context, bucket, key, contentType string, body io.ReadSeeker, size int64) error {
	s.puts = append(s.puts, bucket)
	s.keys = append(s.keys, key)
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if int64(len(data)) != size || contentType == "" {
		return fmt.Errorf("incorrect file metadata")
	}
	if bucket == s.failBucket {
		return errors.New("storage error")
	}
	return nil
}
func (s *fakeStorage) Delete(ctx context.Context, bucket, key string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	s.deletes = append(s.deletes, bucket)
	s.keys = append(s.keys, key)
	if bucket == s.deleteFailBucket {
		return errors.New("delete failed")
	}
	return nil
}

type fakeRepository struct {
	listBooks   []bookrepo.Book
	stored      bookrepo.StoredBook
	input       bookrepo.CreateInput
	getErr      error
	deleteErr   error
	err         error
	listInput   bookrepo.ListInput
	deleteCalls int
	calls       int
}

func (r *fakeRepository) Create(_ context.Context, input bookrepo.CreateInput) (string, error) {
	r.calls++
	r.input = input
	if r.err != nil {
		return "", r.err
	}
	return "book-id", nil
}
func TestUpload(t *testing.T) {
	tests := []struct {
		name       string
		failBucket string
		dbErr      error
		puts       []string
		deletes    []string
		calls      int
		wantErr    bool
	}{
		{name: "success", puts: []string{testBooksBucket, testCoversBucket}, calls: 1},
		{name: "book upload fails", failBucket: testBooksBucket, puts: []string{testBooksBucket}, wantErr: true},
		{name: "cover upload fails", failBucket: testCoversBucket, puts: []string{testBooksBucket, testCoversBucket}, deletes: []string{testBooksBucket}, wantErr: true},
		{name: "database fails", dbErr: errors.New("db error"), puts: []string{testBooksBucket, testCoversBucket}, deletes: []string{testBooksBucket, testCoversBucket}, calls: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &fakeStorage{failBucket: tt.failBucket}
			repo := &fakeRepository{err: tt.dbErr}
			service := NewBookService(repo, storage, testBooksBucket, testCoversBucket)
			input := UploadInput{Title: "Go", Author: "Author", Description: "Description", AccessLevel: "free", Level: "beginner", Topic: "Go", Book: File{Body: bytes.NewReader([]byte("book")), Size: 4, ContentType: testPDFType}, Cover: File{Body: bytes.NewReader([]byte("cover")), Size: 5, ContentType: "image/png"}}
			result, err := service.Upload(context.Background(), input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(storage.puts, tt.puts) || !reflect.DeepEqual(storage.deletes, tt.deletes) || repo.calls != tt.calls {
				t.Fatalf("puts %v, deletes %v, db calls %d", storage.puts, storage.deletes, repo.calls)
			}
			for _, key := range storage.keys {
				if key == "" || key != storage.keys[0] {
					t.Fatalf("inconsistent object keys: %v", storage.keys)
				}
			}
			if !tt.wantErr && !uploadResultMatches(result, input, repo.input) {
				t.Fatalf("unexpected result %+v or repository input %+v", result, repo.input)
			}
		})
	}
}

func (r *fakeRepository) List(_ context.Context, input bookrepo.ListInput) ([]bookrepo.Book, error) {
	r.listInput = input
	return r.listBooks, r.err
}

func TestListPagination(t *testing.T) {
	repo := &fakeRepository{listBooks: []bookrepo.Book{{ID: "1"}, {ID: "2"}, {ID: "3"}}}
	service := NewBookService(repo, &fakeStorage{}, testBooksBucket, testCoversBucket)
	result, err := service.List(context.Background(), ListInput{Limit: 2, Offset: 4, Search: "go", Topic: "Go"})
	if err != nil || len(result.Books) != 2 || !result.HasMore {
		t.Fatalf("result: %+v, error: %v", result, err)
	}
	if repo.listInput.Limit != 3 || repo.listInput.Offset != 4 || repo.listInput.Search != "go" || repo.listInput.Topic != "Go" {
		t.Fatalf("input: %+v", repo.listInput)
	}
}

func (r *fakeRepository) Get(_ context.Context, _ string) (bookrepo.StoredBook, error) {
	return r.stored, r.getErr
}
func (r *fakeRepository) Delete(_ context.Context, _ string) error {
	r.deleteCalls++
	return r.deleteErr
}
func (s *fakeStorage) Get(_ context.Context, bucket, key string) (io.ReadCloser, string, int64, error) {
	return io.NopCloser(bytes.NewReader([]byte("book"))), testPDFType, 4, nil
}

func TestDeleteBook(t *testing.T) {
	for _, tt := range []struct {
		name    string
		fail    string
		dbErr   error
		deletes []string
		calls   int
		wantErr bool
	}{
		{name: "success", deletes: []string{testBooksBucket, testCoversBucket}, calls: 1},
		{name: "book delete fails", fail: testBooksBucket, wantErr: true, deletes: []string{testBooksBucket}},
		{name: "cover delete fails", fail: testCoversBucket, wantErr: true, deletes: []string{testBooksBucket, testCoversBucket}},
		{name: "database delete fails", dbErr: errors.New("db"), wantErr: true, deletes: []string{testBooksBucket, testCoversBucket}, calls: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{stored: bookrepo.StoredBook{BookKey: "bk", CoverKey: "ck"}, deleteErr: tt.dbErr}
			storage := &fakeStorage{deleteFailBucket: tt.fail}
			err := NewBookService(repo, storage, testBooksBucket, testCoversBucket).Delete(context.Background(), "id")
			if (err != nil) != tt.wantErr || !reflect.DeepEqual(storage.deletes, tt.deletes) || repo.deleteCalls != tt.calls {
				t.Fatalf("err: %v, deletes: %v, db calls: %d", err, storage.deletes, repo.deleteCalls)
			}
		})
	}
}

func TestDownloadBook(t *testing.T) {
	repo := &fakeRepository{stored: bookrepo.StoredBook{Title: "Go", BookFilename: "go.pdf", BookKey: "key"}}
	result, err := NewBookService(repo, &fakeStorage{}, testBooksBucket, testCoversBucket).Download(context.Background(), "id")
	if err != nil {
		t.Fatal(err)
	}
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil || string(data) != "book" || result.Filename != "go.pdf" || result.ContentType != testPDFType {
		t.Fatalf("result: %+v; body: %q; err: %v", result, data, err)
	}
}

func uploadResultMatches(result UploadResult, input UploadInput, saved bookrepo.CreateInput) bool {
	return result.ID == "book-id" && result.Title == input.Title && result.BookKey == result.CoverKey && result.BookKey == saved.BookKey && saved.CoverKey == result.BookKey && saved.AccessLevel == input.AccessLevel
}
