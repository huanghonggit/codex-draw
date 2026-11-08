package main

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

func openLegacySecret(s storedSecret) (string, error) {
	if s.Kind != "macos-keychain" {
		return "", errors.New("unsupported legacy storage")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	b, e := exec.CommandContext(ctx, "/usr/bin/security", "find-generic-password", "-a", s.Data, "-s", "codex-image2", "-w").Output()
	if e != nil {
		return "", errors.New("cannot read legacy Keychain item; run configure")
	}
	return strings.TrimSpace(string(b)), nil
}
