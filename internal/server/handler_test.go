package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abhi267266/botnet-practice/pkg/protocol"
	"github.com/abhi267266/botnet-practice/pkg/security"
	"github.com/stretchr/testify/assert"
)

func TestHandleHeartbeat(t *testing.T) {
	registry := NewRegistry()
	secret := "test-secret"
	handler := HandleHeartbeat(registry, secret)

	// Construct Valid Encrypted Request
	reqData := protocol.HeartbeatRequest{AgentID: "agent-123"}
	payload, _ := json.Marshal(reqData)
	
	ciphertext, nonce, err := security.Encrypt(payload, secret)
	assert.NoError(t, err)

	ts := time.Now().Unix()
	sig := security.Sign(ciphertext, nonce, ts, secret)

	env := protocol.EncryptedEnvelope{
		Data:      ciphertext,
		Nonce:     nonce,
		Timestamp: ts,
		Signature: sig,
	}

	envBody, _ := json.Marshal(env)

	req := httptest.NewRequest(http.MethodPost, "/v1/heartbeat", bytes.NewBuffer(envBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify registry was updated
	active := registry.GetActiveAgents(10 * time.Second)
	assert.Contains(t, active, "agent-123")

	// Verify the response is an encrypted envelope
	var respEnv protocol.EncryptedEnvelope
	err = json.Unmarshal(w.Body.Bytes(), &respEnv)
	assert.NoError(t, err)

	// Verify response signature
	err = security.VerifySignature(respEnv.Data, respEnv.Nonce, respEnv.Timestamp, respEnv.Signature, secret)
	assert.NoError(t, err)

	// Decrypt response
	respPayload, err := security.Decrypt(respEnv.Data, respEnv.Nonce, secret)
	assert.NoError(t, err)

	var resp protocol.HeartbeatResponse
	err = json.Unmarshal(respPayload, &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)

	// Invalid Request (Bad Signature)
	badEnv := env
	badEnv.Signature = "bad-sig"
	badBody, _ := json.Marshal(badEnv)
	reqBad := httptest.NewRequest(http.MethodPost, "/v1/heartbeat", bytes.NewBuffer(badBody))
	wBad := httptest.NewRecorder()

	handler.ServeHTTP(wBad, reqBad)
	assert.Equal(t, http.StatusUnauthorized, wBad.Code)
}
