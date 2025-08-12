package log

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer

	// Initialize with debug level and JSON format for easier parsing
	InitWithConfig(Config{
		Level:  Debug,
		Format: JSON,
		Output: &buf,
	})

	// Log at various levels
	slog.Debug("debug message", "key", "value")
	slog.Info("info message", "key", "value")
	slog.Warn("warn message", "key", "value")
	slog.Error("error message", "key", "value")

	// Parse the log entries
	entries := parseJSONLogEntries(t, buf.String())
	if len(entries) != 4 {
		t.Errorf("Expected 4 log entries, got %d", len(entries))
	}

	// Reset buffer
	buf.Reset()

	// Now test with Info level
	InitWithConfig(Config{
		Level:  Info,
		Format: JSON,
		Output: &buf,
	})

	// Log at various levels again
	slog.Debug("debug message", "key", "value") // Should be filtered out
	slog.Info("info message", "key", "value")
	slog.Warn("warn message", "key", "value")
	slog.Error("error message", "key", "value")

	// Parse the log entries
	entries = parseJSONLogEntries(t, buf.String())
	if len(entries) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(entries))
	}
}

func TestLogFormats(t *testing.T) {
	// Test JSON format
	var jsonBuf bytes.Buffer
	InitWithConfig(Config{
		Level:  Info,
		Format: JSON,
		Output: &jsonBuf,
	})

	slog.Info("test message", "key", "value")
	jsonOut := jsonBuf.String()
	if !strings.Contains(jsonOut, `"key":"value"`) {
		t.Errorf("JSON log format doesn't contain expected key-value: %s", jsonOut)
	}

	// Test text format
	var textBuf bytes.Buffer
	InitWithConfig(Config{
		Level:  Info,
		Format: Text,
		Output: &textBuf,
	})

	slog.Info("test message", "key", "value")
	textOut := textBuf.String()
	if !strings.Contains(textOut, "key=value") {
		t.Errorf("Text log format doesn't contain expected key-value: %s", textOut)
	}
}

func TestWithFunctions(t *testing.T) {
	var buf bytes.Buffer
	InitWithConfig(Config{
		Level:  Info,
		Format: JSON,
		Output: &buf,
	})

	// Test With
	logger := With("component", "test")
	logger.Info("with logger")

	entries := parseJSONLogEntries(t, buf.String())
	if len(entries) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(entries))
	}
	entry := entries[0]
	if entry["component"] != "test" {
		t.Errorf("Expected component=test, got component=%v", entry["component"])
	}

	// Reset
	buf.Reset()

	// Test WithGroup
	groupLogger := WithGroup("testgroup")
	groupLogger.Info("group test", "key", "value")

	entries = parseJSONLogEntries(t, buf.String())
	if len(entries) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(entries))
	}
}

// Helper function to parse JSON log entries
func parseJSONLogEntries(t *testing.T, logOutput string) []map[string]any {
	var entries []map[string]any

	for _, line := range strings.Split(strings.TrimSpace(logOutput), "\n") {
		var entry map[string]any
		err := json.Unmarshal([]byte(line), &entry)
		if err != nil {
			t.Fatalf("Failed to parse JSON log entry: %v", err)
		}
		entries = append(entries, entry)
	}

	return entries
}
