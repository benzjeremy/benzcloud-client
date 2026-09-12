package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/benzjeremy/benzcloud-client/internal/pairing"
)

//go:embed web/*
var webFS embed.FS

const Version = "v1.0"

func main() {
	var (
		portFlag    int
		configFlag  string
		serverFlag  string
		userFlag    string
		passFlag    string
		daemonFlag  bool
		versionFlag bool
	)

	flag.IntVar(&portFlag, "port", 8088, "Local client control port")
	flag.StringVar(&configFlag, "config", "", "Client configuration directory")
	flag.StringVar(&serverFlag, "server", "", "Server URL to pair with via CLI")
	flag.StringVar(&userFlag, "user", "", "Username for pairing")
	flag.StringVar(&passFlag, "pass", "", "Password for pairing")
	flag.BoolVar(&daemonFlag, "daemon", false, "Run in background daemon mode")
	flag.BoolVar(&versionFlag, "version", false, "Print version and exit")
	flag.Parse()

	if versionFlag {
		fmt.Printf("BenzCloud Client %s (Lead Engineer: Jeremy Benz • GNU GPLv3)\n", Version)
		return
	}

	if configFlag == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			configFlag = "./benzcloud-client-data"
		} else {
			configFlag = filepath.Join(home, ".benzcloud", "client")
		}
	}

	_ = os.MkdirAll(configFlag, 0700)

	// CLI pairing mode
	if serverFlag != "" && userFlag != "" && passFlag != "" {
		fmt.Printf("==> Pairing with BenzCloud Server at %s as %s...\n", serverFlag, userFlag)
		prof, err := pairing.Pair(serverFlag, userFlag, passFlag, configFlag)
		if err != nil {
			log.Fatalf("Pairing failed: %v\n", err)
		}
		fmt.Printf("✅ Success! Paired with %s (Overlay IP: %s, Base Domain: %s)\n",
			prof.ServerLANURL, prof.OverlayIP, prof.BaseDomain)
		return
	}

	// Client local control HTTP server
	mux := http.NewServeMux()

	mux.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		prof, err := pairing.Load(configFlag)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"paired": false})
			return
		}
		_ = json.NewEncoder(w).Encode(prof)
	})

	mux.HandleFunc("/api/pair", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ServerURL string `json:"server_url"`
			Username  string `json:"username"`
			Password  string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		prof, err := pairing.Pair(req.ServerURL, req.Username, req.Password, configFlag)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(prof)
	})

	mux.HandleFunc("/api/unpair", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = pairing.Unpair(configFlag)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})

	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Fatal: failed to load embedded web assets: %v\n", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", portFlag),
		Handler: mux,
	}

	appURL := fmt.Sprintf("http://127.0.0.1:%d", portFlag)

	go func() {
		log.Printf("🚀 [BenzCloud Client %s] Running on: %s\n", Version, appURL)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Fatal: client server failed: %v\n", err)
		}
	}()

	if !daemonFlag && os.Getenv("DISPLAY") != "" && os.Getenv("HEADLESS") != "1" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			LaunchGUI("BenzCloud Client", appURL, 900, 680)
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\nStopping BenzCloud Client...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	log.Println("BenzCloud Client safely exited.")
}
