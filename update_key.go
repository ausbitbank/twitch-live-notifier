package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

// updateSigningPublicKeyHex optionally pins an Ed25519 public key for
// update-manifest.json. Leave empty to accept unsigned manifests; releases
// without signing secrets therefore keep working.
var updateSigningPublicKeyHex = ""

func updateSignatureRequired() bool {
	return strings.TrimSpace(updateSigningPublicKeyHex) != ""
}

func verifyUpdateManifestSignature(manifest, signature []byte) error {
	if !updateSignatureRequired() {
		return nil
	}
	key, err := hex.DecodeString(strings.TrimSpace(updateSigningPublicKeyHex))
	if err != nil || len(key) != ed25519.PublicKeySize {
		return errors.New("the update signing key is invalid")
	}
	raw := signature
	if len(raw) != ed25519.SignatureSize {
		trimmed := bytes.TrimSpace(signature)
		if decoded, decodeErr := hex.DecodeString(string(trimmed)); decodeErr == nil && len(decoded) == ed25519.SignatureSize {
			raw = decoded
		} else if decoded, decodeErr := base64.StdEncoding.DecodeString(string(trimmed)); decodeErr == nil && len(decoded) == ed25519.SignatureSize {
			raw = decoded
		}
	}
	if len(raw) != ed25519.SignatureSize {
		return errors.New("the update manifest signature is invalid")
	}
	if !ed25519.Verify(ed25519.PublicKey(key), manifest, raw) {
		return errors.New("the update manifest signature does not match")
	}
	return nil
}
