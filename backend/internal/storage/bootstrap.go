package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BootstrapBuckets creates missing global buckets and grants the application key access.
func BootstrapBuckets(ctx context.Context, endpoint, token, key string, names ...string) error {
	if token == "" || key == "" {
		return fmt.Errorf("GARAGE_ADMIN_TOKEN and GARAGE_DEFAULT_ACCESS_KEY are required")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	call := func(method, path string, body any, out any) error {
		var b bytes.Buffer
		if body != nil {
			if err := json.NewEncoder(&b).Encode(body); err != nil {
				return err
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint+path, &b)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			return fmt.Errorf("garage admin %s: %s", path, res.Status)
		}
		if out != nil {
			return json.NewDecoder(res.Body).Decode(out)
		}
		return nil
	}
	var buckets []struct {
		ID      string   `json:"id"`
		Aliases []string `json:"globalAliases"`
	}
	if err := call("GET", "/v2/ListBuckets", nil, &buckets); err != nil {
		return err
	}
	for _, name := range names {
		id := ""
		for _, bucket := range buckets {
			for _, alias := range bucket.Aliases {
				if alias == name {
					id = bucket.ID
				}
			}
		}
		if id == "" {
			var created struct {
				ID string `json:"id"`
			}
			if err := call("POST", "/v2/CreateBucket", map[string]string{"globalAlias": name}, &created); err != nil {
				return err
			}
			id = created.ID
		}
		if id == "" {
			return fmt.Errorf("garage returned empty bucket id for %s", name)
		}
		permissions := map[string]any{"accessKeyId": key, "bucketId": id, "permissions": map[string]bool{"read": true, "write": true, "owner": false}}
		if err := call("POST", "/v2/AllowBucketKey", permissions, nil); err != nil {
			return err
		}
	}
	return nil
}
