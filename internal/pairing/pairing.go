package pairing

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ClientProfile stores the pairing status and credentials.
type ClientProfile struct {
	ServerLANURL  string            `json:"server_lan_url"`
	Username      string            `json:"username"`
	BaseDomain    string            `json:"base_domain"`
	OverlayIP     string            `json:"overlay_ip"`
	ServerVPNIP   string            `json:"server_vpn_ip"`
	SessionToken  string            `json:"session_token"`
	Paired        bool              `json:"paired"`
	PairedAt      time.Time         `json:"paired_at"`
	Permissions   map[string]bool   `json:"permissions"`
	ConfigDir     string            `json:"config_dir"`
	mu            sync.RWMutex
}

var ErrNotPaired = errors.New("client is not paired with a BenzCloud server")

// Pair connects to the BenzCloud server via local IP, exchanges credentials, and saves the mesh VPN config.
func Pair(serverURL, username, password, configDir string) (*ClientProfile, error) {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	cleanURL := strings.TrimRight(serverURL, "/")
	pairEndpoint := cleanURL + "/api/pair"

	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(pairEndpoint, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("pairing request to %s failed: %w", pairEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("server rejected pairing (HTTP %d): %s", resp.StatusCode, errResp["error"])
	}

	var pairData struct {
		Status           string          `json:"status"`
		Username         string          `json:"username"`
		BaseDomain       string          `json:"base_domain"`
		OverlayIP        string          `json:"overlay_ip"`
		ServerVPNIP      string          `json:"server_vpn_ip"`
		CertPEM          string          `json:"cert_pem"`
		KeyPEM           string          `json:"key_pem"`
		ClientConfigYAML string          `json:"client_config_yaml"`
		SessionToken     string          `json:"session_token"`
		Permissions      map[string]bool `json:"permissions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pairData); err != nil {
		return nil, fmt.Errorf("failed to parse pairing response: %w", err)
	}

	// Write mesh VPN files to disk
	yamlPath := filepath.Join(configDir, "client_nebula.yaml")
	if err := os.WriteFile(yamlPath, []byte(pairData.ClientConfigYAML), 0600); err != nil {
		return nil, fmt.Errorf("failed to write nebula config: %w", err)
	}

	profile := &ClientProfile{
		ServerLANURL: cleanURL,
		Username:     pairData.Username,
		BaseDomain:   pairData.BaseDomain,
		OverlayIP:    pairData.OverlayIP,
		ServerVPNIP:  pairData.ServerVPNIP,
		SessionToken: pairData.SessionToken,
		Paired:       true,
		PairedAt:     time.Now().UTC(),
		Permissions:  pairData.Permissions,
		ConfigDir:    configDir,
	}

	if err := profile.Save(); err != nil {
		return nil, err
	}
	return profile, nil
}

// ProfilePath returns the location of pairing.json.
func (p *ClientProfile) ProfilePath() string {
	return filepath.Join(p.ConfigDir, "pairing.json")
}

// Save persists the profile to disk with 0600 permissions.
func (p *ClientProfile) Save() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmpFile := p.ProfilePath() + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, p.ProfilePath())
}

// Load loads the profile from configDir.
func Load(configDir string) (*ClientProfile, error) {
	pPath := filepath.Join(configDir, "pairing.json")
	data, err := os.ReadFile(pPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClientProfile{ConfigDir: configDir, Paired: false}, ErrNotPaired
		}
		return nil, err
	}

	var p ClientProfile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	p.ConfigDir = configDir
	return &p, nil
}

// Unpair removes pairing data from disk.
func Unpair(configDir string) error {
	pPath := filepath.Join(configDir, "pairing.json")
	yamlPath := filepath.Join(configDir, "client_nebula.yaml")
	_ = os.Remove(yamlPath)
	return os.Remove(pPath)
}
