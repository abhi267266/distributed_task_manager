package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrTimestampExpired = errors.New("timestamp expired")
)

// Encrypt encrypts the plaintext using AES-256-GCM. 
// The secret is hashed with SHA256 to create a 32-byte key.
func Encrypt(plaintext []byte, secret string) ([]byte, []byte, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts the ciphertext using AES-256-GCM and the provided nonce.
func Decrypt(ciphertext, nonce []byte, secret string) ([]byte, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

// Sign generates an HMAC-SHA256 signature over data, nonce, and timestamp.
func Sign(data, nonce []byte, timestamp int64, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	mac.Write(nonce)
	mac.Write([]byte(fmt.Sprintf("%d", timestamp)))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature checks if the provided signature matches the calculated one.
func VerifySignature(data, nonce []byte, timestamp int64, signature, secret string) error {
	expected := Sign(data, nonce, timestamp, secret)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return ErrInvalidSignature
	}
	return nil
}

// ValidateTimestamp checks if a timestamp is within the allowed window to prevent replay attacks.
func ValidateTimestamp(ts int64, window time.Duration) error {
	msgTime := time.Unix(ts, 0)
	now := time.Now()
	
	diff := now.Sub(msgTime)
	if diff < -window || diff > window {
		return ErrTimestampExpired
	}
	return nil
}
