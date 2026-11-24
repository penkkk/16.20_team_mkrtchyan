package book

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
)

func makeEPUB(t *testing.T, valid bool) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	header := &zip.FileHeader{Name: "mimetype", Method: zip.Store}
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	kind := epubMIME
	if !valid {
		kind = "application/not-epub"
	}
	if _, writeErr := w.Write([]byte(kind)); writeErr != nil {
		t.Fatal(writeErr)
	}
	w, err = zw.Create("META-INF/container.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("<container/>")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestUploadChecksContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pdf := []byte("%PDF-1.7\nexample")
	epub := makeEPUB(t, true)
	png := []byte("\x89PNG\r\n\x1a\nmore-data")
	jpeg := []byte("\xff\xd8\xff\xe0more-data")
	webp := []byte("RIFF\x10\x00\x00\x00WEBPVP8 more-data")
	for _, tt := range []struct {
		name                string
		bookType, coverType string
		book, cover         []byte
		status              int
	}{
		{"pdf and png", pdfMIME, "image/png", pdf, png, 201},
		{"epub and jpeg", epubMIME, "image/jpeg", epub, jpeg, 201},
		{"pdf and webp", pdfMIME, "image/webp", pdf, webp, 201},
		{"fake pdf", "", "", []byte("not a pdf"), png, 400},
		{"zip without epub mime", "", "", makeEPUB(t, false), png, 400},
		{"truncated epub", "", "", epub[:len(epub)/2], png, 400},
		{"fake cover", "", "", pdf, []byte("not an image"), 400},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			for key, val := range map[string]string{"title": "title", "author": "author", "description": "description", "topic": "go", "level": "beginner", "access_level": "free"} {
				if err := form.WriteField(key, val); err != nil {
					t.Fatal(err)
				}
			}
			for _, file := range []struct {
				name, filename string
				data           []byte
			}{{"book", "book.pdf", tt.book}, {"cover", "cover.png", tt.cover}} {
				hdr := make(textproto.MIMEHeader)
				hdr.Set("Content-Disposition", `form-data; name="`+file.name+`"; filename="`+file.filename+`"`)
				hdr.Set("Content-Type", "application/incorrect")
				part, err := form.CreatePart(hdr)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := part.Write(file.data); err != nil {
					t.Fatal(err)
				}
			}
			if err := form.Close(); err != nil {
				t.Fatal(err)
			}
			svc := &fakeService{}
			router := gin.New()
			pass := func(c *gin.Context) { c.Next() }
			NewHandler(svc).RegisterRoutes(router.Group("/api/v1"), pass, pass)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/upload", &body)
			request.Header.Set("Content-Type", form.FormDataContentType())
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if tt.status == 400 && svc.calls != 0 {
				t.Fatalf("invalid upload reached service")
			}
			if tt.status == 201 && !uploadMatches(svc, tt.book, tt.cover, tt.bookType, tt.coverType) {
				t.Fatalf("invalid file data or content types: %+v", svc.uploadInput)
			}
		})
	}
}

func uploadMatches(svc *fakeService, book, cover []byte, bookType, coverType string) bool {
	return svc.calls == 1 && svc.uploadInput.Book.ContentType == bookType && svc.uploadInput.Cover.ContentType == coverType && bytes.Equal(svc.bookBytes, book) && bytes.Equal(svc.coverBytes, cover)
}
