package tests

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/url"
	"testing"

	authservice "opd/internal/service/auth"

	"github.com/redis/go-redis/v9"
)

type captureRedisSet struct {
	key  string
	data []byte
}

func (h *captureRedisSet) DialHook(next redis.DialHook) redis.DialHook { return next }
func (h *captureRedisSet) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (h *captureRedisSet) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if cmd.Name() != "set" {
			return next(ctx, cmd)
		}
		args := cmd.Args()
		h.key = args[1].(string)
		h.data = args[2].([]byte)
		cmd.(*redis.StatusCmd).SetVal("OK")
		return nil
	}
}

func TestStartOAuthStoresSafeReturnURLAndPKCE(t *testing.T) {
	for _, tc := range []struct {
		input, want string
		link        bool
	}{
		{"/profile?tab=books", "/profile?tab=books", false},
		{"https://app.example.com/profile", "/profile", false},
		{"https://evil.example.com/profile", "/", false},
		{"//evil.example.com/profile", "/", false},
		{"javascript:alert(1)", "/", false},
		{"/settings", "/settings", true},
	} {
		t.Run(tc.input, func(t *testing.T) {
			hook := &captureRedisSet{}
			client := redis.NewClient(&redis.Options{Addr: "unused:0", Dialer: func(context.Context, string, string) (net.Conn, error) {
				t.Fatal("unexpected Redis connection")
				return nil, nil
			}})
			client.AddHook(hook)
			t.Cleanup(func() { _ = client.Close() })
			service := authservice.NewAuthService(nil, nil, nil, client, authservice.OAuthConfig{
				GoogleClientID: "client-id", AppPublicURL: "https://app.example.com",
			})
			var result authservice.StartOAuthResult
			var err error
			if tc.link {
				result, err = service.StartOAuthLink(context.Background(), authservice.StartOAuthLinkInput{Provider: "google", ReturnURL: tc.input, UserID: "user-1"})
			} else {
				result, err = service.StartOAuth(context.Background(), authservice.StartOAuthInput{Provider: "google", ReturnURL: tc.input})
			}
			if err != nil {
				t.Fatal(err)
			}
			var attempt authservice.OAuthAttempt
			if err := json.Unmarshal(hook.data, &attempt); err != nil {
				t.Fatal(err)
			}
			wantMode, wantUserID := authservice.OAuthModeLogin, ""
			if tc.link {
				wantMode, wantUserID = authservice.OAuthModeLink, "user-1"
			}
			if hook.key != authservice.OAuthRedisKey+result.State || attempt.ReturnURL != tc.want || attempt.Provider != "google" || attempt.Mode != wantMode || attempt.UserID != wantUserID || attempt.CodeVerifier == "" || attempt.Nonce == "" {
				t.Fatalf("unexpected stored OAuth attempt: key=%q, attempt=%+v", hook.key, attempt)
			}
			u, err := url.Parse(result.RedirectURL)
			if err != nil {
				t.Fatal(err)
			}
			challenge := sha256.Sum256([]byte(attempt.CodeVerifier))
			if q := u.Query(); q.Get("state") != result.State || q.Get("nonce") != attempt.Nonce ||
				q.Get("code_challenge") != base64.RawURLEncoding.EncodeToString(challenge[:]) ||
				q.Get("redirect_uri") != "https://app.example.com/api/v1/auth/google/callback" {
				t.Fatalf("unexpected authorization URL: %s", result.RedirectURL)
			}
		})
	}
}
