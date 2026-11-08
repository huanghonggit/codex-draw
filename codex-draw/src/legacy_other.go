//go:build !windows && !darwin

package main

import "errors"

func openLegacySecret(s storedSecret) (string, error) {
	return "", errors.New("legacy migration supports Windows and macOS only")
}
