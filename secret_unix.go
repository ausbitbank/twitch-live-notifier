//go:build !windows

package main

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const secretKeyringService = "TwitchLiveNotifier"
const secretKeyringAccount = "master-key"

// keyringDisabled lets tests and headless installs avoid D-Bus/Keychain
// prompts. When set, the fallback 0600 key file is used.
func keyringDisabled() bool {
	value := strings.TrimSpace(os.Getenv("TWITCH_LIVE_DISABLE_KEYRING"))
	return value == "1" || strings.EqualFold(value, "true")
}

var keyringMasterOnce sync.Once
var keyringMasterKey []byte
var keyringMasterOK bool

func newPlatformSecretProtector(dir string) (secretProtector, error) {
	if !keyringDisabled() {
		keyringMasterOnce.Do(func() {
			if key, ok := keyringLookupKey(); ok {
				keyringMasterKey = key
				keyringMasterOK = true
				return
			}
			key, err := randomSecretKey()
			if err == nil && keyringStoreKey(key) {
				keyringMasterKey = key
				keyringMasterOK = true
			}
		})
		if keyringMasterOK {
			return newAESGCMProtector(keyringMasterKey)
		}
	}
	return newKeyFileProtector(dir)
}

func keyringLookupKey() ([]byte, bool) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return nil, false
	}
	args := keyringCommand("lookup")
	if args == nil {
		return nil, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(output)))
	if err != nil || len(key) != secretKeySize {
		return nil, false
	}
	return key, true
}

func keyringStoreKey(key []byte) bool {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return false
	}
	command := keyringCommand("store")
	if command == nil {
		return false
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	if runtime.GOOS == "darwin" {
		command = append(command, encoded)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	if runtime.GOOS != "darwin" {
		cmd.Stdin = strings.NewReader(encoded)
	}
	return cmd.Run() == nil
}

func keyringCommand(action string) []string {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("secret-tool"); err != nil {
			return nil
		}
		if action == "lookup" {
			return []string{"secret-tool", "lookup", "service", secretKeyringService, "account", secretKeyringAccount}
		}
		return []string{"secret-tool", "store", "--label=" + secretKeyringService, "service", secretKeyringService, "account", secretKeyringAccount}
	case "darwin":
		if _, err := exec.LookPath("security"); err != nil {
			return nil
		}
		if action == "lookup" {
			return []string{"security", "find-generic-password", "-s", secretKeyringService, "-a", secretKeyringAccount, "-w"}
		}
		return []string{"security", "add-generic-password", "-U", "-s", secretKeyringService, "-a", secretKeyringAccount, "-w"}
	default:
		return nil
	}
}
