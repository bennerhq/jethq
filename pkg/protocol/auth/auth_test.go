package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientDisablesKeepAlives(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := client.HTTPClient().Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected HTTP client transport to be *http.Transport")
	}
	if !transport.DisableKeepAlives {
		t.Fatal("expected keep-alives to be disabled")
	}
}

func TestLoginReturnsDeviceErrorMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid password"}`))
	}))
	defer srv.Close()

	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	err = client.Login(context.Background(), srv.URL, "wrong")
	if err == nil || err.Error() != "Invalid password" {
		t.Fatalf("login error = %v, want Invalid password", err)
	}
	var authErr *Error
	if !errors.As(err, &authErr) {
		t.Fatal("expected login error to unwrap to *auth.Error")
	}
	if authErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d", authErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestLoginReturnsRetryAfterMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"Too many failed attempts. Please try again later.","retry_after":312}`))
	}))
	defer srv.Close()

	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	err = client.Login(context.Background(), srv.URL, "wrong")
	want := "Too many failed attempts. Please try again later. (retry after 312s)"
	if err == nil || err.Error() != want {
		t.Fatalf("login error = %v, want %q", err, want)
	}
	var authErr *Error
	if !errors.As(err, &authErr) {
		t.Fatal("expected login error to unwrap to *auth.Error")
	}
	if authErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status code = %d, want %d", authErr.StatusCode, http.StatusTooManyRequests)
	}
}

func TestGetDeviceInfoParsesLocalAuthMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device" {
			t.Fatalf("path = %q, want /device", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"authMode":     "password",
			"deviceId":     "jetkvm-test",
			"loopbackOnly": true,
		})
	}))
	defer srv.Close()

	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	info, err := client.GetDeviceInfo(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if info.AuthMode != LocalAuthModePassword {
		t.Fatalf("auth mode = %v, want password", info.AuthMode)
	}
	if info.DeviceID != "jetkvm-test" {
		t.Fatalf("device id = %q, want jetkvm-test", info.DeviceID)
	}
	if !info.LoopbackOnly {
		t.Fatal("expected loopbackOnly true")
	}
}

func TestLocalPasswordMutationsUseExpectedEndpoints(t *testing.T) {
	type requestRecord struct {
		Method string
		Path   string
	}
	var requests []requestRecord
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, requestRecord{Method: r.Method, Path: r.URL.Path})
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/auth/password-local":
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPut && r.URL.Path == "/auth/password-local":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/auth/local-password":
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CreateLocalPassword(context.Background(), srv.URL, "password123"); err != nil {
		t.Fatal(err)
	}
	if err := client.UpdateLocalPassword(context.Background(), srv.URL, "password123", "password456"); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteLocalPassword(context.Background(), srv.URL, "password456"); err != nil {
		t.Fatal(err)
	}

	want := []requestRecord{
		{Method: http.MethodPost, Path: "/auth/password-local"},
		{Method: http.MethodPut, Path: "/auth/password-local"},
		{Method: http.MethodDelete, Path: "/auth/local-password"},
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %#v, want %#v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request %d = %#v, want %#v", i, requests[i], want[i])
		}
	}
}
