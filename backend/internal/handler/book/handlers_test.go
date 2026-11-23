package book

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	bookservice "opd/internal/service/book"
)

type fakeService struct {
	input       bookservice.ListInput
	uploadInput bookservice.UploadInput
	bookBytes   []byte
	coverBytes  []byte
	calls       int
}

func (s *fakeService) Upload(_ context.Context, input bookservice.UploadInput) (bookservice.UploadResult, error) {
	s.calls++
	s.uploadInput = input
	s.bookBytes, _ = io.ReadAll(input.Book.Body)
	s.coverBytes, _ = io.ReadAll(input.Cover.Body)
	return bookservice.UploadResult{}, nil
}
func (s *fakeService) List(_ context.Context, input bookservice.ListInput) (bookservice.ListResult, error) {
	s.calls++
	s.input = input
	return bookservice.ListResult{Books: []bookservice.Book{{ID: "123", Title: "Go", AccessLevel: "free", Level: "beginner", Topic: "Go"}}, HasMore: true}, nil
}
func TestListRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name, url  string
		authStatus int
		status     int
		calls      int
	}{
		{"valid", "/api/v1/admin/books?limit=10&offset=20&search=go&topic=Go", 0, 200, 1},
		{"invalid limit", "/api/v1/admin/books?limit=101", 0, 400, 0},
		{"invalid offset", "/api/v1/admin/books?offset=-1", 0, 400, 0},
		{"unauthorized", "/api/v1/admin/books", 401, 401, 0},
		{"not admin", "/api/v1/admin/books", 403, 403, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{}
			router := gin.New()
			auth := func(c *gin.Context) {
				if tt.authStatus == 401 {
					c.AbortWithStatus(401)
					return
				}
				c.Next()
			}
			admin := func(c *gin.Context) {
				if tt.authStatus == 403 {
					c.AbortWithStatus(403)
					return
				}
				c.Next()
			}
			NewHandler(svc).RegisterRoutes(router.Group("/api/v1"), auth, admin)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.url, nil))
			if recorder.Code != tt.status || svc.calls != tt.calls {
				t.Fatalf("status %d, calls %d", recorder.Code, svc.calls)
			}
			if tt.name == "valid" {
				if svc.input.Limit != 10 || svc.input.Offset != 20 || svc.input.Search != "go" || svc.input.Topic != "Go" {
					t.Fatalf("input: %+v", svc.input)
				}
				var result ListResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil || len(result.Books) != 1 || !result.HasMore || result.Books[0].Title != "Go" {
					t.Fatalf("response: %s, error: %v", recorder.Body.String(), err)
				}
			}
		})
	}
}

func (s *fakeService) Delete(context.Context, string) error { s.calls++; return nil }
func (s *fakeService) Download(context.Context, string) (bookservice.DownloadResult, error) {
	s.calls++
	return bookservice.DownloadResult{Body: io.NopCloser(strings.NewReader("content")), ContentType: "application/pdf", Size: 7, Filename: "example.pdf"}, nil
}

func TestBookActions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := "01234567-89ab-cdef-0123-456789abcdef"
	for _, tt := range []struct {
		method, path string
		denied       bool
		status       int
		calls        int
	}{
		{"GET", "/books/" + id + "/download", false, 200, 1},
		{"DELETE", "/books/" + id, false, 204, 1},
		{"GET", "/books/invalid/download", false, 400, 0},
		{"DELETE", "/books/" + id, true, 403, 0},
		{"GET", "/books/" + id + "/download", true, 403, 0},
	} {
		svc := &fakeService{}
		router := gin.New()
		pass := func(c *gin.Context) { c.Next() }
		admin := func(c *gin.Context) {
			if tt.denied {
				c.AbortWithStatus(403)
				return
			}
			c.Next()
		}
		NewHandler(svc).RegisterRoutes(router.Group("/admin-test"), pass, admin)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(tt.method, "/admin-test/admin"+tt.path, nil))
		if w.Code != tt.status || svc.calls != tt.calls {
			t.Fatalf("%s %s: status %d, calls %d", tt.method, tt.path, w.Code, svc.calls)
		}
		if tt.method == "GET" && tt.status == 200 && (w.Body.String() != "content" || !strings.Contains(w.Header().Get("Content-Disposition"), "example.pdf")) {
			t.Fatalf("bad download response: %v %q", w.Header(), w.Body.String())
		}
	}
}
