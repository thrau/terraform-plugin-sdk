package retry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetStateChangeConfOverrides(t *testing.T) {
	// Since GetStateChangeConfOverrides uses sync.Once, and we can't reset it
	// without export hacks, we'll test the internal loadOverrides logic
	// by assuming GetStateChangeConfOverrides calls it once.
	// We can still test that it returns something.

	overrides := GetStateChangeConfOverrides()
	if _, ok := overrides["sqs.waitQueueDeleted"]; !ok {
		t.Errorf("expected default override sqs.waitQueueDeleted to be present")
	}
}

func TestLoadOverrides(t *testing.T) {
	oldWd, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "test-retry")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	content := `{
		"file.override": {
			"Delay": 5
		}
	}`
	err = os.WriteFile(filepath.Join(tmpDir, overrideFileName), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Directly call loadOverrides to test its logic without sync.Once
	loadOverrides()

	if _, ok := cachedOverrides["file.override"]; !ok {
		t.Errorf("expected file.override to be present in cachedOverrides")
	}

	if cachedOverrides["file.override"].Delay != 5*time.Second {
		t.Errorf("expected Delay 5s, got %s", cachedOverrides["file.override"].Delay)
	}

	// Ensure default overrides are still there
	if _, ok := cachedOverrides["sqs.waitQueueDeleted"]; !ok {
		t.Errorf("expected default override sqs.waitQueueDeleted to be present")
	}
}

func TestLoadFromConfig(t *testing.T) {
	content := `{
		"test.override": {
			"Delay": 1,
			"Pending": ["pending"],
			"Target": ["done"],
			"Timeout": 10,
			"MinTimeout": 2,
			"PollInterval": 3,
			"NotFoundChecks": 5,
			"ContinuousTargetOccurence": 3
		}
	}`
	tmpfile, err := os.CreateTemp("", "config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	overrides, err := LoadFromConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %s", err)
	}

	override, ok := overrides["test.override"]
	if !ok {
		t.Fatal("expected test.override to be present")
	}

	if override.Delay != 1*time.Second {
		t.Errorf("expected Delay 1s, got %s", override.Delay)
	}
	if len(override.Pending) != 1 || override.Pending[0] != "pending" {
		t.Errorf("expected Pending [pending], got %v", override.Pending)
	}
	if len(override.Target) != 1 || override.Target[0] != "done" {
		t.Errorf("expected Target [done], got %v", override.Target)
	}
	if override.Timeout != 10*time.Second {
		t.Errorf("expected Timeout 10s, got %s", override.Timeout)
	}
	if override.MinTimeout != 2*time.Second {
		t.Errorf("expected MinTimeout 2s, got %s", override.MinTimeout)
	}
	if override.PollInterval != 3*time.Second {
		t.Errorf("expected PollInterval 3s, got %s", override.PollInterval)
	}
	if override.NotFoundChecks != 5 {
		t.Errorf("expected NotFoundChecks 5, got %d", override.NotFoundChecks)
	}
	if override.ContinuousTargetOccurence != 3 {
		t.Errorf("expected ContinuousTargetOccurence 3, got %d", override.ContinuousTargetOccurence)
	}
}
