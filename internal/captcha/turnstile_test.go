package captcha

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"crm_lite/internal/core/config"
)

func TestVerifyTurnstile(t *testing.T) {
	// 设置测试配置
	testOpts := &config.Options{
		Auth: config.AuthOptions{
			Captcha: config.CaptchaOptions{
				TurnstileSecret: "test-secret",
			},
		},
	}
	config.SetInstanceForTest(testOpts)

	tests := []struct {
		name      string
		handler   http.HandlerFunc
		expectOK  bool
		expectErr bool
	}{
		{
			name:      "network failure",
			handler:   nil, // we'll shut down server to simulate
			expectOK:  false,
			expectErr: true,
		},
		{
			name:      "non 2xx status",
			handler:   func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) },
			expectOK:  false,
			expectErr: true,
		},
		{
			name:      "malformed JSON",
			handler:   func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("{invalid")) },
			expectOK:  false,
			expectErr: true,
		},
		{
			name:      "success false",
			handler:   func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"success":false}`)) },
			expectOK:  false,
			expectErr: false,
		},
		{
			name:      "success true",
			handler:   func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"success":true}`)) },
			expectOK:  true,
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// For network failure case, we assign an invalid URL.
			if tc.handler == nil {
				turnstileURL = "http://127.0.0.1:1" // unreachable
			} else {
				srv := httptest.NewServer(tc.handler)
				defer srv.Close()
				turnstileURL = srv.URL
			}

			ok, err := VerifyTurnstile(context.Background(), "dummy-token", "")
			if tc.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok != tc.expectOK {
				t.Fatalf("expected ok=%v, got %v", tc.expectOK, ok)
			}
		})
	}
}
