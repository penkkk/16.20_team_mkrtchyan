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
			json.NewEncoder(w).Encode([]any{map[string]any{"id": "books-id", "globalAliases": []string{"books"}}})
		case "/v2/CreateBucket":
			var request map[string]string
			json.NewDecoder(r.Body).Decode(&request)
			if request["globalAlias"] != "covers" {
				t.Errorf("unexpected bucket: %v", request)
			}
			created++
			json.NewEncoder(w).Encode(map[string]string{"id": "covers-id"})
		case "/v2/AllowBucketKey":
			var request struct {
				AccessKeyID string          `json:"accessKeyId"`
				BucketID    string          `json:"bucketId"`
				Permissions map[string]bool `json:"permissions"`
			}
			json.NewDecoder(r.Body).Decode(&request)
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
