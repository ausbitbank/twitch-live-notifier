package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyUpdateManifestSignatureOptional(t *testing.T) {
	old := updateSigningPublicKeyHex
	t.Cleanup(func() { updateSigningPublicKeyHex = old })

	updateSigningPublicKeyHex = ""
	if err := verifyUpdateManifestSignature([]byte("manifest"), nil); err != nil {
		t.Fatalf("unsigned manifests should be accepted when no key is set: %v", err)
	}

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	updateSigningPublicKeyHex = hex.EncodeToString(public)
	manifest := []byte(`{"version":"2.6.0"}`)
	signature := ed25519.Sign(private, manifest)
	if err := verifyUpdateManifestSignature(manifest, signature); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	if err := verifyUpdateManifestSignature([]byte(`{"version":"2.6.1"}`), signature); err == nil {
		t.Fatal("tampered manifest accepted")
	}
}

func TestPinnedKeyRejectsUnsignedManifestWithoutFallback(t *testing.T) {
	old := updateSigningPublicKeyHex
	t.Cleanup(func() { updateSigningPublicKeyHex = old })

	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	updateSigningPublicKeyHex = hex.EncodeToString(public)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"9.9.9","assets":{"linux-amd64":{"name":"TwitchLiveNotifier-linux-amd64","url":"https://example.test/bin","sha256":"` + strings.Repeat("a", 64) + `","size":10}}}`))
	}))
	defer server.Close()

	release := githubRelease{
		TagName: "v9.9.9",
		Assets: []githubAsset{
			{Name: updateManifestAssetName, BrowserDownloadURL: server.URL + "/manifest"},
			{Name: updateChecksumsAssetName, BrowserDownloadURL: server.URL + "/sums"},
		},
	}
	if _, err := resolveReleaseUpdateAsset(context.Background(), release, installTarget{Mode: installModeInPlace}, "linux", "amd64"); err == nil {
		t.Fatal("expected an unsigned manifest to fail closed when a key is pinned")
	}
}

func TestPinnedKeyAcceptsSignedManifest(t *testing.T) {
	old := updateSigningPublicKeyHex
	t.Cleanup(func() { updateSigningPublicKeyHex = old })

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	updateSigningPublicKeyHex = hex.EncodeToString(public)
	manifest := []byte(`{"version":"9.9.9","assets":{"linux-amd64":{"name":"TwitchLiveNotifier-linux-amd64","url":"https://example.test/bin","sha256":"` + strings.Repeat("a", 64) + `","size":10}}}`)
	signature := ed25519.Sign(private, manifest)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/signature":
			_, _ = w.Write(signature)
		default:
			_, _ = w.Write(manifest)
		}
	}))
	defer server.Close()

	release := githubRelease{
		TagName: "v9.9.9",
		Assets: []githubAsset{
			{Name: updateManifestAssetName, BrowserDownloadURL: server.URL + "/manifest"},
			{Name: updateManifestSignatureName, BrowserDownloadURL: server.URL + "/signature"},
		},
	}
	asset, err := resolveReleaseUpdateAsset(context.Background(), release, installTarget{Mode: installModeInPlace}, "linux", "amd64")
	if err != nil {
		t.Fatalf("signed manifest rejected: %v", err)
	}
	if asset.Name != "TwitchLiveNotifier-linux-amd64" || asset.SHA256 == "" {
		t.Fatalf("unexpected asset: %+v", asset)
	}
}
