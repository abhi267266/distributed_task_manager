package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/abhi267266/botnet-practice/internal/server"
	"github.com/abhi267266/botnet-practice/pkg/common"
)

func main() {
	common.InitLogger()
	slog.Info("Starting Server...")

	var cfg common.ServerConfig
	err := common.LoadConfig("server-config.yaml", &cfg)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("Server configuration loaded", "port", cfg.Port, "address", cfg.Address)

	// Level 1 - Basic Server Listener
	registry := server.NewRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/heartbeat", server.HandleHeartbeat(registry, cfg.SharedSecret))

	addr := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	slog.Info("Starting HTTPS server", "address", addr, "cert", cfg.CertFile)
	
	if err := http.ListenAndServeTLS(addr, cfg.CertFile, cfg.KeyFile, mux); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
