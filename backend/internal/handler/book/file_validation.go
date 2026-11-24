package book

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"
)

const (
	pdfMIME  = "application/pdf"
	epubMIME = "application/epub+zip"
)

var errUnsupportedFile = errors.New("unsupported or invalid file type")

func validateBook(file io.ReaderAt, size int64) (string, error) {
	header := make([]byte, 5)
	if _, err := file.ReadAt(header, 0); err == nil && bytes.Equal(header, []byte("%PDF-")) {
		return pdfMIME, nil
	}
	archive, err := zip.NewReader(file, size)
	if err != nil {
		return "", errUnsupportedFile
	}
	return validateEPUB(archive)
}

func validateEPUB(archive *zip.Reader) (string, error) {
	var mimetype *zip.File
	hasContainer := false
	for _, entry := range archive.File {
		switch entry.Name {
		case "mimetype":
			if mimetype != nil {
				return "", errUnsupportedFile
			}
			mimetype = entry
		case "META-INF/container.xml":
			hasContainer = true
		}
	}
	if mimetype == nil || mimetype.Method != zip.Store || mimetype.UncompressedSize64 != uint64(len(epubMIME)) || !hasContainer {
		return "", errUnsupportedFile
	}
	r, err := mimetype.Open()
	if err != nil {
		return "", errUnsupportedFile
	}
	data, err := io.ReadAll(io.LimitReader(r, int64(len(epubMIME)+1)))
	closeErr := r.Close()
	if err != nil || closeErr != nil || string(data) != epubMIME {
		return "", errUnsupportedFile
	}
	return epubMIME, nil
}

func validateCover(file io.ReaderAt, size int64) (string, error) {
	header := make([]byte, min(size, 512))
	n, err := file.ReadAt(header, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	contentType := http.DetectContentType(header[:n])
	switch contentType {
	case "image/png", "image/jpeg", "image/webp":
		return contentType, nil
	default:
		return "", errUnsupportedFile
	}
}
