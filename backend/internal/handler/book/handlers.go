package book

import (
	"errors"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	bookrepo "opd/internal/repository/book"
	bookservice "opd/internal/service/book"

	"github.com/gin-gonic/gin"
)

const (
	maxUploadSize = 100 << 20
	errorKey      = "error"
	uploadFailed  = "upload failed"
)

func (h *Handler) upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	if err := c.Request.ParseMultipartForm(8 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid multipart form or upload exceeds 100 MiB"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() {
			if err := c.Request.MultipartForm.RemoveAll(); err != nil {
				log.Printf("remove multipart files: %v", err)
			}
		}()
	}
	req := UploadRequest{
		Title:       strings.TrimSpace(c.PostForm("title")),
		Author:      strings.TrimSpace(c.PostForm("author")),
		Description: strings.TrimSpace(c.PostForm("description")),
		Topic:       strings.TrimSpace(c.PostForm("topic")),
		AccessLevel: c.PostForm("access_level"),
		Level:       c.PostForm("level"),
	}
	title, author := req.Title, req.Author
	description, topic := req.Description, req.Topic
	access, level := req.AccessLevel, req.Level
	if !validUploadRequest(req) {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "title, author, description, topic, valid access_level and level are required"})
		return
	}
	bookFile, err := c.FormFile("book")
	if err != nil || bookFile.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "book file is required"})
		return
	}
	coverFile, err := c.FormFile("cover")
	if err != nil || coverFile.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "cover file is required"})
		return
	}
	bookBody, err := bookFile.Open()
	if err != nil {
		log.Printf("open book: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: uploadFailed})
		return
	}
	defer bookBody.Close()
	coverBody, err := coverFile.Open()
	if err != nil {
		log.Printf("open cover: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: uploadFailed})
		return
	}
	defer coverBody.Close()
	bookType, err := validateBook(bookBody, bookFile.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "book must be a PDF or EPUB file"})
		return
	}
	coverType, err := validateCover(coverBody, coverFile.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "cover must be a PNG, JPEG or WebP image"})
		return
	}
	result, err := h.service.Upload(c.Request.Context(), bookservice.UploadInput{
		Title: title, Author: author, Description: description, AccessLevel: access, Level: level, Topic: topic,
		Book:  bookservice.File{Filename: safeFilename(bookFile.Filename), Body: bookBody, Size: bookFile.Size, ContentType: bookType},
		Cover: bookservice.File{Body: coverBody, Size: coverFile.Size, ContentType: coverType},
	})
	if err != nil {
		log.Printf("upload book: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: uploadFailed})
		return
	}
	c.JSON(http.StatusCreated, UploadResponse{
		ID: result.ID, Title: result.Title, Author: result.Author, Description: result.Description,
		AccessLevel: result.AccessLevel, Level: result.Level, Topic: result.Topic,
		BookKey: result.BookKey, CoverKey: result.CoverKey,
	})
}

func (h *Handler) list(c *gin.Context) {
	limit, offset := 20, 0
	var err error
	if raw := c.Query("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "limit must be between 1 and 100"})
			return
		}
	}
	if raw := c.Query("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 || offset > 2147483647 {
			c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid offset"})
			return
		}
	}
	result, err := h.service.List(c.Request.Context(), bookservice.ListInput{Limit: limit, Offset: offset, Search: strings.TrimSpace(c.Query("search")), Topic: strings.TrimSpace(c.Query("topic"))})
	if err != nil {
		log.Printf("list books: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "could not load books"})
		return
	}
	books := make([]BookResponse, 0, len(result.Books))
	for _, book := range result.Books {
		books = append(books, BookResponse{ID: book.ID, Title: book.Title, Author: book.Author, Description: book.Description, AccessLevel: book.AccessLevel, Level: book.Level, Topic: book.Topic})
	}
	c.JSON(http.StatusOK, ListResponse{Books: books, Limit: limit, Offset: offset, HasMore: result.HasMore})
}

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (h *Handler) download(c *gin.Context) {
	id := c.Param("id")
	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid book id"})
		return
	}
	result, err := h.service.Download(c.Request.Context(), id)
	if err != nil {
		respondBookError(c, err)
		return
	}
	defer result.Body.Close()
	filename := downloadFilename(result.Title, result.Filename, result.ContentType)
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": filename})
	if disposition == "" {
		disposition = `attachment; filename="book"`
	}
	c.Header("Content-Disposition", disposition)
	c.Header("X-Content-Type-Options", "nosniff")
	c.DataFromReader(http.StatusOK, result.Size, "application/octet-stream", result.Body, nil)
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid book id"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondBookError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondBookError(c *gin.Context, err error) {
	if errors.Is(err, bookrepo.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{errorKey: "book not found"})
		return
	}
	log.Printf("book operation: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{errorKey: "book operation failed"})
}

func safeFilename(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, `\`, "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if name == "." || name == ".." || name == "/" {
		return ""
	}
	return name
}

func validUploadRequest(req UploadRequest) bool {
	return req.Title != "" && req.Author != "" && req.Description != "" && req.Topic != "" &&
		(req.AccessLevel == "free" || req.AccessLevel == "subscription") &&
		(req.Level == "beginner" || req.Level == "intermediate" || req.Level == "advanced")
}

func downloadFilename(title, original, contentType string) string {
	name := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, title)
	name = strings.Trim(strings.TrimSpace(name), ". ")
	if name == "" {
		name = "book"
	}
	var extension string
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case pdfMIME:
		extension = ".pdf"
	case epubMIME:
		extension = ".epub"
	default:
		candidate := strings.ToLower(filepath.Ext(safeFilename(original)))
		if candidate == ".pdf" || candidate == ".epub" {
			extension = candidate
		}
	}
	if extension != "" && !strings.HasSuffix(strings.ToLower(name), extension) {
		name += extension
	}
	return name
}
