package book

import "io"

type File struct {
	Body        io.ReadSeeker
	Filename    string
	ContentType string
	Size        int64
}

type UploadInput struct {
	Title       string
	Author      string
	Description string
	AccessLevel string
	Level       string
	Topic       string
	Book        File
	Cover       File
}

type UploadResult struct {
	ID          string
	Title       string
	Author      string
	Description string
	AccessLevel string
	Level       string
	Topic       string
	BookKey     string
	CoverKey    string
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

type ListResult struct {
	Books   []Book
	HasMore bool
}

type DownloadResult struct {
	Body        io.ReadCloser
	ContentType string
	Filename    string
	Title       string
	Size        int64
}
