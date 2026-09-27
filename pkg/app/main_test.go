package app

import (
	"fmt"
	"os"
	"testing"
)

// TestMain redirects the preferences file into a throwaway home for the whole
// test binary. Connecting and scanning both persist devices, so without this
// any test that builds an App writes into the developer's real profile and
// litters the launcher list with loopback test servers.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "jethq-test-home")
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating test home: %v\n", err)
		os.Exit(1)
	}
	userHomeDir = func() (string, error) { return home, nil }

	code := m.Run()

	_ = os.RemoveAll(home)
	os.Exit(code)
}
