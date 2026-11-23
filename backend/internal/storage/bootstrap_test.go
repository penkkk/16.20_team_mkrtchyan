package storage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBootstrapCreatesMissingBucketAndGrantsAccess(t *testing.T) {
	var created int
	granted := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Error("missing admin token")
		}
		switch r.URL.Path {
		case "/v2/ListBuckets":
			if err := json.NewEncoder(w).Encode([]any{map[string]any{"id": "books-id", "globalAliases": []string{"books"}}}); err != nil {
				t.Error(err)
			}
		case "/v2/CreateBucket":
			var request map[string]string
			decodeRequest(t, r, &request)
			if request["globalAlias"] != "covers" {
				t.Errorf("unexpected bucket: %v", request)
			}
			created++
			if err := json.NewEncoder(w).Encode(map[string]string{"id": "covers-id"}); err != nil {
				t.Error(err)
			}
		case "/v2/AllowBucketKey":
			var request struct {
				Permissions map[string]bool `json:"permissions"`
				AccessKeyID string          `json:"accessKeyId"`
				BucketID    string          `json:"bucketId"`
			}
			decodeRequest(t, r, &request)
			if request.AccessKeyID != "key" || !request.Permissions["read"] || !request.Permissions["write"] {
				t.Errorf("unexpected grant: %+v", request)
			}
			granted[request.BucketID] = true
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	if err := BootstrapBuckets(context.Background(), server.URL, "token", "key", "books", "covers"); err != nil {
		t.Fatal(err)
	}
	if created != 1 || !granted["books-id"] || !granted["covers-id"] {
		t.Fatalf("created %d; grants %v", created, granted)
	}
}

func decodeRequest(t *testing.T, r *http.Request, dst any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		t.Error(err)
	}
}
