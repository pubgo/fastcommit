//go:build !darwin

package main

import "errors"

func authenticateWithBiometricsOrPasscode(_ string) error {
	return errors.New("device owner authentication is only available on macOS")
}
