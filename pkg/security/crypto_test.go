package security

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	secret := "test-secret"
	plaintext := []byte("hello botnet")

	ciphertext, nonce, err := Encrypt(plaintext, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEmpty(t, nonce)
	
	// Should not equal plaintext
	assert.False(t, bytes.Equal(ciphertext, plaintext))

	// Decrypt
	decrypted, err := Decrypt(ciphertext, nonce, secret)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// Decrypt with wrong secret
	_, err = Decrypt(ciphertext, nonce, "wrong-secret")
	assert.Error(t, err)
}

func TestSignature(t *testing.T) {
	secret := "test-secret"
	data := []byte("some-data")
	nonce := []byte("random-nonce")
	timestamp := time.Now().Unix()

	sig := Sign(data, nonce, timestamp, secret)
	assert.NotEmpty(t, sig)

	// Verify valid
	err := VerifySignature(data, nonce, timestamp, sig, secret)
	assert.NoError(t, err)

	// Verify tampered data
	err = VerifySignature([]byte("tampered"), nonce, timestamp, sig, secret)
	assert.ErrorIs(t, err, ErrInvalidSignature)

	// Verify tampered timestamp
	err = VerifySignature(data, nonce, timestamp+1, sig, secret)
	assert.ErrorIs(t, err, ErrInvalidSignature)
}

func TestValidateTimestamp(t *testing.T) {
	window := 30 * time.Second

	// Valid (now)
	err := ValidateTimestamp(time.Now().Unix(), window)
	assert.NoError(t, err)

	// Valid (20 seconds ago)
	err = ValidateTimestamp(time.Now().Add(-20*time.Second).Unix(), window)
	assert.NoError(t, err)

	// Invalid (40 seconds ago)
	err = ValidateTimestamp(time.Now().Add(-40*time.Second).Unix(), window)
	assert.ErrorIs(t, err, ErrTimestampExpired)
	
	// Invalid (40 seconds in future - e.g. bad clock sync or malicious)
	err = ValidateTimestamp(time.Now().Add(40*time.Second).Unix(), window)
	assert.ErrorIs(t, err, ErrTimestampExpired)
}
