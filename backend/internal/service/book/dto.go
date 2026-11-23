package book

import "io"

type File struct {
	Filename    string
	Body        io.ReadSeeker
	Size        int64
	ContentType string
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
	Limit  int
	Offset int
	Search string
	Topic  string
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
	Size        int64
	Filename    string
	Title       string
}
