package book

type UploadRequest struct {
	Title       string
	Author      string
	Description string
	AccessLevel string
	Level       string
	Topic       string
}

type UploadResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	AccessLevel string `json:"access_level"`
	Level       string `json:"level"`
	Topic       string `json:"topic"`
	BookKey     string `json:"book_key"`
	CoverKey    string `json:"cover_key"`
}

type BookResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	AccessLevel string `json:"access_level"`
	Level       string `json:"level"`
	Topic       string `json:"topic"`
}

type ListResponse struct {
	Books   []BookResponse `json:"books"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
}
