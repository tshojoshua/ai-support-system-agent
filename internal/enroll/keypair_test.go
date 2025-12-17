package enroll

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	if kp == nil {
		t.Fatal("Expected non-nil keypair")
	}

	if len(kp.PublicKey) != ed25519.PublicKeySize {
		t.Errorf("Public key size: got %d, want %d", len(kp.PublicKey), ed25519.PublicKeySize)
	}

	if len(kp.PrivateKey) != ed25519.PrivateKeySize {
		t.Errorf("Private key size: got %d, want %d", len(kp.PrivateKey), ed25519.PrivateKeySize)
	}
}

func TestKeyPairBase64Encoding(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// Encode to base64
	pubB64 := kp.PublicKeyBase64()
	privB64 := kp.PrivateKeyBase64()

	// Should be valid base64
	if _, err := base64.StdEncoding.DecodeString(pubB64); err != nil {
		t.Errorf("Public key not valid base64: %v", err)
	}

	if _, err := base64.StdEncoding.DecodeString(privB64); err != nil {
		t.Errorf("Private key not valid base64: %v", err)
	}
}

func TestLoadKeyPair(t *testing.T) {
	// Generate original keypair
	original, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// Encode to base64
	pubB64 := original.PublicKeyBase64()
	privB64 := original.PrivateKeyBase64()

	// Load from base64
	loaded, err := LoadKeyPair(pubB64, privB64)
	if err != nil {
		t.Fatalf("LoadKeyPair() error = %v", err)
	}

	// Compare keys
	if string(loaded.PublicKey) != string(original.PublicKey) {
		t.Error("Public keys don't match")
	}

	if string(loaded.PrivateKey) != string(original.PrivateKey) {
		t.Error("Private keys don't match")
	}
}

func TestLoadKeyPairInvalidBase64(t *testing.T) {
	tests := []struct {
		name   string
		pubKey string
		privKey string
	}{
		{
			name:   "invalid public key",
			pubKey: "not-valid-base64!@#$",
			privKey: base64.StdEncoding.EncodeToString(make([]byte, ed25519.PrivateKeySize)),
		},
		{
			name:   "invalid private key",
			pubKey: base64.StdEncoding.EncodeToString(make([]byte, ed25519.PublicKeySize)),
			privKey: "not-valid-base64!@#$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadKeyPair(tt.pubKey, tt.privKey)
			if err == nil {
				t.Error("Expected error for invalid base64")
			}
		})
	}
}

func TestLoadKeyPairInvalidSize(t *testing.T) {
	tests := []struct {
		name   string
		pubSize int
		privSize int
	}{
		{
			name:   "invalid public key size",
			pubSize: 16, // Too small
			privSize: ed25519.PrivateKeySize,
		},
		{
			name:   "invalid private key size",
			pubSize: ed25519.PublicKeySize,
			privSize: 32, // Too small
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubB64 := base64.StdEncoding.EncodeToString(make([]byte, tt.pubSize))
			privB64 := base64.StdEncoding.EncodeToString(make([]byte, tt.privSize))

			_, err := LoadKeyPair(pubB64, privB64)
			if err == nil {
				t.Error("Expected error for invalid key size")
			}
		})
	}
}
