package protocol

// HeartbeatRequest is sent by the agent to the server to signal it is alive.
type HeartbeatRequest struct {
	AgentID string `json:"agent_id"`
}

// HeartbeatResponse is the server's reply to a HeartbeatRequest.
type HeartbeatResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// EncryptedEnvelope is used to wrap all secure communications.
type EncryptedEnvelope struct {
	Data      []byte `json:"data"`      // Encrypted payload
	Nonce     []byte `json:"nonce"`     // AES GCM nonce
	Timestamp int64  `json:"timestamp"` // Unix timestamp for replay protection
	Signature string `json:"signature"` // HMAC-SHA256 signature
}
