package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	const loadedKey = "PROTOCOL_PARSER_TEST_LOADED"
	const preservedKey = "PROTOCOL_PARSER_TEST_PRESERVED"
	os.Unsetenv(loadedKey)
	t.Cleanup(func() { os.Unsetenv(loadedKey) })
	t.Setenv(preservedKey, "system-value")

	path := filepath.Join(t.TempDir(), ".env")
	content := "# database settings\n" + loadedKey + "=file-value\n" + preservedKey + "=file-value\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LoadEnvFile(path); err != nil {
		t.Fatal(err)
	}
	if os.Getenv(loadedKey) != "file-value" {
		t.Fatalf("expected file value, got %q", os.Getenv(loadedKey))
	}
	if os.Getenv(preservedKey) != "system-value" {
		t.Fatalf("system environment must win, got %q", os.Getenv(preservedKey))
	}
}
