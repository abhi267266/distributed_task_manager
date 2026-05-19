package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/abhi267266/botnet-practice/internal/agent"
	"github.com/abhi267266/botnet-practice/pkg/common"
)

func main() {
	common.InitLogger()
	slog.Info("Starting Agent...")

	var cfg common.AgentConfig
	err := common.LoadConfig("agent-config.yaml", &cfg)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("Agent configuration loaded", "server_url", cfg.ServerURL, "interval", cfg.Interval)

	// Level 1 - Connectivity & Identity
	agentID, err := agent.LoadOrGenerateID(".agent_id")
	if err != nil {
		slog.Error("Failed to initialize agent identity", "error", err)
		os.Exit(1)
	}

	// Setup context that cancels on interrupt (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		slog.Info("Shutting down agent...")
		cancel()
	}()

	agent.StartHeartbeat(ctx, cfg.ServerURL, agentID, cfg.Interval, cfg.SharedSecret, cfg.ServerCertFile)
	slog.Info("Agent stopped.")
}
