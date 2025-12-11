package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPortFromFile_Success(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	portFile := filepath.Join(tmpDir, "forwarded_port")

	// Write a valid port
	err := os.WriteFile(portFile, []byte("12345"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	w := &Watcher{portFile: portFile}
	port, err := w.readPortFromFile()
	if err != nil {
		t.Fatalf("readPortFromFile() error = %v, want nil", err)
	}
	if port != 12345 {
		t.Errorf("readPortFromFile() = %d, want 12345", port)
	}
}

func TestReadPortFromFile_WithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	portFile := filepath.Join(tmpDir, "forwarded_port")

	// Write a port with whitespace
	err := os.WriteFile(portFile, []byte("  54321  \n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	w := &Watcher{portFile: portFile}
	port, err := w.readPortFromFile()
	if err != nil {
		t.Fatalf("readPortFromFile() error = %v, want nil", err)
	}
	if port != 54321 {
		t.Errorf("readPortFromFile() = %d, want 54321", port)
	}
}

func TestReadPortFromFile_FileNotFound(t *testing.T) {
	w := &Watcher{portFile: "/nonexistent/path/port"}
	_, err := w.readPortFromFile()
	if err == nil {
		t.Error("readPortFromFile() error = nil, want error")
	}
}

func TestReadPortFromFile_InvalidPort(t *testing.T) {
	tmpDir := t.TempDir()
	portFile := filepath.Join(tmpDir, "forwarded_port")

	// Write invalid port content
	err := os.WriteFile(portFile, []byte("not-a-number"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	w := &Watcher{portFile: portFile}
	_, err = w.readPortFromFile()
	if err == nil {
		t.Error("readPortFromFile() error = nil, want error")
	}
}

func TestReadPortFromFile_PortOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{"port too low", "0"},
		{"port too high", "65536"},
		{"negative port", "-1"},
		{"very high port", "99999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			portFile := filepath.Join(tmpDir, "forwarded_port")

			err := os.WriteFile(portFile, []byte(tt.port), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			w := &Watcher{portFile: portFile}
			_, err = w.readPortFromFile()
			if err == nil {
				t.Errorf("readPortFromFile() with port %s: error = nil, want error", tt.port)
			}
		})
	}
}

func TestReadPortFromFile_ValidEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected int
	}{
		{"minimum valid port", "1", 1},
		{"maximum valid port", "65535", 65535},
		{"common port", "8080", 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			portFile := filepath.Join(tmpDir, "forwarded_port")

			err := os.WriteFile(portFile, []byte(tt.port), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			w := &Watcher{portFile: portFile}
			port, err := w.readPortFromFile()
			if err != nil {
				t.Errorf("readPortFromFile() error = %v, want nil", err)
			}
			if port != tt.expected {
				t.Errorf("readPortFromFile() = %d, want %d", port, tt.expected)
			}
		})
	}
}
