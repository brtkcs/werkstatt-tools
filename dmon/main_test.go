package main

import (
	"os"
	"path/filepath"
	"testing"
)

func init() {
	cOlive, cRust, cMuted, cReset = "", "", "", ""
}

// mockChecker returns predefined results for process checks
type mockChecker struct {
	status map[string]bool
}

func (m mockChecker) isRunning(name string) bool {
	return m.status[name]
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	yaml := `interval: 10
processes:
  - sshd
  - podman
`
	path := writeTempFile(t, "dmon.yaml", yaml)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Interval != 10 {
		t.Errorf("got interval %d, want 10", cfg.Interval)
	}
	if len(cfg.Processes) != 2 {
		t.Fatalf("got %d processes, want 2", len(cfg.Processes))
	}
	if cfg.Processes[0] != "sshd" {
		t.Errorf("got %q, want sshd", cfg.Processes[0])
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := loadConfig("/nonexistent/dmon.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadConfigDefaultInterval(t *testing.T) {
	yaml := `processes:
  - sshd
`
	path := writeTempFile(t, "dmon.yaml", yaml)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Interval != 5 {
		t.Errorf("got interval %d, want 5 (default)", cfg.Interval)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	path := writeTempFile(t, "dmon.yaml", "{{invalid")
	_, err := loadConfig(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid",
			cfg:  Config{Interval: 5, Processes: []string{"sshd"}},
		},
		{
			name:    "no processes",
			cfg:     Config{Interval: 5, Processes: []string{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCheckAllFirstRun(t *testing.T) {
	chk := mockChecker{status: map[string]bool{
		"sshd":   true,
		"podman": true,
		"nginx":  false,
	}}

	prev := make(map[string]bool)
	changes := checkAll([]string{"sshd", "podman", "nginx"}, prev, chk)

	if len(changes) != 3 {
		t.Fatalf("got %d changes, want 3", len(changes))
	}

	for _, c := range changes {
		if !c.First {
			t.Errorf("%s should be marked as first run", c.Name)
		}
		if c.Changed {
			t.Errorf("%s should not be marked as changed on first run", c.Name)
		}
	}

	// sshd and podman running, nginx not
	if !changes[0].Running {
		t.Error("sshd should be running")
	}
	if !changes[1].Running {
		t.Error("podman should be running")
	}
	if changes[2].Running {
		t.Error("nginx should not be running")
	}
}

func TestCheckAllDetectsChanges(t *testing.T) {
	chk := mockChecker{status: map[string]bool{
		"sshd":  true,
		"nginx": true, // was false, now true
	}}

	prev := map[string]bool{
		"sshd":  true,  // no change
		"nginx": false, // changed!
	}

	changes := checkAll([]string{"sshd", "nginx"}, prev, chk)

	// sshd: no change
	if changes[0].Changed {
		t.Error("sshd should not have changed")
	}
	if changes[0].First {
		t.Error("sshd should not be first")
	}

	// nginx: changed from false to true
	if !changes[1].Changed {
		t.Error("nginx should have changed")
	}
	if !changes[1].Running {
		t.Error("nginx should be running now")
	}
}

func TestCheckAllNoChangeNoOutput(t *testing.T) {
	chk := mockChecker{status: map[string]bool{
		"sshd": true,
	}}

	prev := map[string]bool{
		"sshd": true,
	}

	changes := checkAll([]string{"sshd"}, prev, chk)

	if changes[0].First {
		t.Error("should not be first")
	}
	if changes[0].Changed {
		t.Error("should not have changed")
	}
}

func TestCheckAllStopDetection(t *testing.T) {
	chk := mockChecker{status: map[string]bool{
		"sshd": false, // was true, now false
	}}

	prev := map[string]bool{
		"sshd": true,
	}

	changes := checkAll([]string{"sshd"}, prev, chk)

	if !changes[0].Changed {
		t.Error("sshd should have changed")
	}
	if changes[0].Running {
		t.Error("sshd should be stopped")
	}
}
