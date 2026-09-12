package pairing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestClientPairing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "benzcloud_client_pairing_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Mock server /api/pair endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pair" {
			http.NotFound(w, r)
			return
		}
		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req["username"] != "testuser" || req["password"] != "ValidPassword123" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
			return
		}

		resp := map[string]interface{}{
			"status":             "paired",
			"username":           "testuser",
			"base_domain":        "benzjeremy.de",
			"overlay_ip":         "10.42.0.2",
			"server_vpn_ip":      "10.42.0.1",
			"client_config_yaml": "# Mock Client Nebula Config\nlisten: 0.0.0.0:0\n",
			"session_token":      "mock_session_token_12345",
			"permissions": map[string]bool{
				"vpn":   true,
				"dns":   true,
				"drive": true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// 1. Initial Load: should return ErrNotPaired
	_, err = Load(tempDir)
	if err != ErrNotPaired {
		t.Fatalf("Expected ErrNotPaired, got %v", err)
	}

	// 2. Perform Pairing
	profile, err := Pair(mockServer.URL, "testuser", "ValidPassword123", tempDir)
	if err != nil {
		t.Fatalf("Pair failed: %v", err)
	}
	if !profile.Paired || profile.BaseDomain != "benzjeremy.de" || profile.OverlayIP != "10.42.0.2" {
		t.Fatalf("Unexpected profile state: %+v", profile)
	}

	// 3. Reload from disk
	loaded, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.SessionToken != "mock_session_token_12345" {
		t.Fatalf("Loaded token mismatch: %s", loaded.SessionToken)
	}

	// 4. Unpair
	if err := Unpair(tempDir); err != nil {
		t.Fatalf("Unpair failed: %v", err)
	}
	_, err = Load(tempDir)
	if err != ErrNotPaired {
		t.Fatal("Expected ErrNotPaired after Unpair")
	}
}
