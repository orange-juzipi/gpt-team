package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"gpt-team-api/internal/apperr"
)

type Cipher struct {
	aead cipher.AEAD
}

func NewCipher(secret string) (*Cipher, error) {
	key := []byte(secret)
	if len(key) != 32 {
		decoded, err := base64.StdEncoding.DecodeString(secret)
		if err == nil {
			key = decoded
		}
	}

	if len(key) != 32 {
		return nil, apperr.BadRequest("invalid_encryption_key", "ACCOUNT_ENCRYPTION_KEY must be 32 bytes or base64 encoded 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, apperr.Internal("cipher_init_failed", "failed to initialize cipher", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, apperr.Internal("cipher_gcm_failed", "failed to initialize gcm", err)
	}

	return &Cipher{aead: aead}, nil
}

func (c *Cipher) Encrypt(plainText string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", apperr.Internal("cipher_nonce_failed", "failed to generate nonce", err)
	}

	encrypted := c.aead.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (c *Cipher) Decrypt(cipherText string) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", apperr.Internal("cipher_decode_failed", "failed to decode cipher text", err)
	}

	nonceSize := c.aead.NonceSize()
	if len(payload) < nonceSize {
		return "", apperr.Internal("cipher_payload_failed", "cipher payload is invalid", nil)
	}

	nonce := payload[:nonceSize]
	encrypted := payload[nonceSize:]
	plain, err := c.aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", apperr.Internal("cipher_decrypt_failed", "failed to decrypt account password", err)
	}

	return string(plain), nil
}
