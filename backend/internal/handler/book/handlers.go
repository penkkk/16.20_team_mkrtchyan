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

const maxUploadSize = 100 << 20

func (h *Handler) upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	if err := c.Request.ParseMultipartForm(8 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form or upload exceeds 100 MiB"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
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
	if title == "" || author == "" || description == "" || topic == "" || (access != "free" && access != "subscription") || (level != "beginner" && level != "intermediate" && level != "advanced") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, author, description, topic, valid access_level and level are required"})
		return
	}
	bookFile, err := c.FormFile("book")
	if err != nil || bookFile.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "book file is required"})
		return
	}
	coverFile, err := c.FormFile("cover")
	if err != nil || coverFile.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover file is required"})
		return
	}
	bookBody, err := bookFile.Open()
	if err != nil {
		log.Printf("open book: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	defer bookBody.Close()
	coverBody, err := coverFile.Open()
	if err != nil {
		log.Printf("open cover: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}
	defer coverBody.Close()
	bookType, err := validateBook(bookBody, bookFile.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "book must be a PDF or EPUB file"})
		return
	}
	coverType, err := validateCover(coverBody, coverFile.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover must be a PNG, JPEG or WebP image"})
		return
	}
	result, err := h.service.Upload(c.Request.Context(), bookservice.UploadInput{
		Title: title, Author: author, Description: description, AccessLevel: access, Level: level, Topic: topic,
		Book:  bookservice.File{Filename: safeFilename(bookFile.Filename), Body: bookBody, Size: bookFile.Size, ContentType: bookType},
		Cover: bookservice.File{Body: coverBody, Size: coverFile.Size, ContentType: coverType},
	})
	if err != nil {
		log.Printf("upload book: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return
		}
	}
	if raw := c.Query("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 || offset > 2147483647 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
	}
	result, err := h.service.List(c.Request.Context(), bookservice.ListInput{Limit: limit, Offset: offset, Search: strings.TrimSpace(c.Query("search")), Topic: strings.TrimSpace(c.Query("topic"))})
	if err != nil {
		log.Printf("list books: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load books"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}
	result, err := h.service.Download(c.Request.Context(), id)
	if err != nil {
		respondBookError(c, err)
		return
	}
	defer result.Body.Close()
	filename := safeFilename(result.Filename)
	if filename == "" {
		filename = safeFilename(result.Title)
		if filename == "" {
			filename = "book"
		}
		switch strings.ToLower(strings.Split(result.ContentType, ";")[0]) {
		case "application/pdf":
			filename += ".pdf"
		case "application/epub+zip":
			filename += ".epub"
		}
	}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	log.Printf("book operation: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "book operation failed"})
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
