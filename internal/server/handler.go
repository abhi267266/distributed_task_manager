package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/abhi267266/botnet-practice/pkg/protocol"
	"github.com/abhi267266/botnet-practice/pkg/security"
)

// HandleHeartbeat returns an http.HandlerFunc that processes heartbeat requests.
func HandleHeartbeat(registry *Registry, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var env protocol.EncryptedEnvelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			slog.Warn("Failed to decode envelope", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify Replay Protection (e.g., 30s window)
		if err := security.ValidateTimestamp(env.Timestamp, 30*time.Second); err != nil {
			slog.Warn("Timestamp validation failed", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify Signature
		if err := security.VerifySignature(env.Data, env.Nonce, env.Timestamp, env.Signature, secret); err != nil {
			slog.Warn("Invalid signature", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Decrypt Payload
		payload, err := security.Decrypt(env.Data, env.Nonce, secret)
		if err != nil {
			slog.Warn("Failed to decrypt payload", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var req protocol.HeartbeatRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			slog.Warn("Failed to unmarshal heartbeat request", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if req.AgentID == "" {
			http.Error(w, "Missing agent_id", http.StatusBadRequest)
			return
		}

		// Update the registry
		registry.UpdateHeartbeat(req.AgentID)
		slog.Debug("Heartbeat received", "agent_id", req.AgentID)

		// Send success response
		resp := protocol.HeartbeatResponse{
			Status: "ok",
		}
		
		respPayload, _ := json.Marshal(resp)
		ciphertext, nonce, err := security.Encrypt(respPayload, secret)
		if err != nil {
			slog.Error("Failed to encrypt response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ts := time.Now().Unix()
		sig := security.Sign(ciphertext, nonce, ts, secret)
		respEnv := protocol.EncryptedEnvelope{
			Data:      ciphertext,
			Nonce:     nonce,
			Timestamp: ts,
			Signature: sig,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respEnv); err != nil {
			slog.Error("Failed to encode heartbeat response", "error", err)
		}
	}
}
