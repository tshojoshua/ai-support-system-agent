package enroll

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// KeyPair represents an Ed25519 keypair
type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateKeyPair creates a new Ed25519 keypair
func GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}
	
	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

// PublicKeyBase64 returns the base64-encoded public key
func (kp *KeyPair) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PublicKey)
}

// PrivateKeyBase64 returns the base64-encoded private key
func (kp *KeyPair) PrivateKeyBase64() string {
	return base64.StdEncoding.EncodeToString(kp.PrivateKey)
}

// LoadKeyPair loads a keypair from base64-encoded strings
func LoadKeyPair(publicKeyB64, privateKeyB64 string) (*KeyPair, error) {
	pub, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	
	priv, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	
	if len(pub) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}
	
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}
	
	return &KeyPair{
		PublicKey:  ed25519.PublicKey(pub),
		PrivateKey: ed25519.PrivateKey(priv),
	}, nil
}
