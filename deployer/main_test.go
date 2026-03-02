package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	cOlive, cRust, cPurple, cMuted, cReset = "", "", "", "", ""
}

// mockRunner simulates command execution for testing
type mockRunner struct {
	output string
	err    error
}

func (m mockRunner) run(runtime, dir string, args ...string) (string, error) {
	return m.output, m.err
}

func (m mockRunner) runAttached(runtime, dir string, args ...string) error {
	return m.err
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
	yaml := `runtime: podman
stacks:
  - name: vaultwarden
    path: /opt/stacks/vaultwarden
  - name: gitea
    path: /opt/stacks/gitea
`
	path := writeTempFile(t, "deployer.yaml", yaml)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Runtime != "podman" {
		t.Errorf("got runtime %q, want podman", cfg.Runtime)
	}
	if len(cfg.Stacks) != 2 {
		t.Fatalf("got %d stacks, want 2", len(cfg.Stacks))
	}
	if cfg.Stacks[0].Name != "vaultwarden" {
		t.Errorf("got name %q, want vaultwarden", cfg.Stacks[0].Name)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := loadConfig("/nonexistent/deployer.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadConfigDefaultRuntime(t *testing.T) {
	yaml := `stacks:
  - name: test
    path: /tmp
`
	path := writeTempFile(t, "deployer.yaml", yaml)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Runtime != "podman" {
		t.Errorf("got runtime %q, want podman (default)", cfg.Runtime)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	path := writeTempFile(t, "deployer.yaml", "{{invalid yaml")
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
			name: "valid config",
			cfg: Config{
				Runtime: "podman",
				Stacks:  []Stack{{Name: "test", Path: "/tmp"}},
			},
		},
		{
			name:    "no stacks",
			cfg:     Config{Runtime: "podman", Stacks: []Stack{}},
			wantErr: true,
		},
		{
			name: "missing name",
			cfg: Config{
				Runtime: "podman",
				Stacks:  []Stack{{Name: "", Path: "/tmp"}},
			},
			wantErr: true,
		},
		{
			name: "missing path",
			cfg: Config{
				Runtime: "podman",
				Stacks:  []Stack{{Name: "test", Path: ""}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFilterStacks(t *testing.T) {
	stacks := []Stack{
		{Name: "vaultwarden", Path: "/opt/vaultwarden"},
		{Name: "gitea", Path: "/opt/gitea"},
		{Name: "immich", Path: "/opt/immich"},
	}

	t.Run("no filter returns all", func(t *testing.T) {
		result, err := filterStacks(stacks, "")
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 3 {
			t.Errorf("got %d stacks, want 3", len(result))
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		result, err := filterStacks(stacks, "gitea")
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 1 || result[0].Name != "gitea" {
			t.Errorf("got %v, want [gitea]", result)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := filterStacks(stacks, "nonexistent")
		if err == nil {
			t.Error("expected error for missing stack")
		}
	})
}

func TestStackStatusRunning(t *testing.T) {
	mock := mockRunner{output: `[{"Name":"vaultwarden","State":"running"}]`, err: nil}
	status, running := stackStatus(Stack{Name: "test", Path: "/tmp"}, "podman", mock)

	if !running {
		t.Error("expected running=true")
	}
	if status != "running" {
		t.Errorf("got status %q, want running", status)
	}
}

func TestStackStatusStopped(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"empty", ""},
		{"empty array", "[]"},
		{"null", "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockRunner{output: tt.output, err: nil}
			status, running := stackStatus(Stack{Name: "test", Path: "/tmp"}, "podman", mock)

			if running {
				t.Error("expected running=false")
			}
			if status != "stopped" {
				t.Errorf("got status %q, want stopped", status)
			}
		})
	}
}

func TestStackStatusError(t *testing.T) {
	mock := mockRunner{output: "", err: fmt.Errorf("command failed")}
	status, running := stackStatus(Stack{Name: "test", Path: "/tmp"}, "podman", mock)

	if running {
		t.Error("expected running=false on error")
	}
	if status != "error" {
		t.Errorf("got status %q, want error", status)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	yaml := `runtime: podman
stacks:
  - name: test
    path: /tmp
`
	path := writeTempFile(t, "deployer.yaml", yaml)
	code := run(path, "invalid", "", mockRunner{})
	if code != 2 {
		t.Errorf("got exit code %d, want 2", code)
	}
}

func TestRunMissingConfig(t *testing.T) {
	code := run("/nonexistent/deployer.yaml", "status", "", mockRunner{})
	if code != 2 {
		t.Errorf("got exit code %d, want 2", code)
	}
}

func TestRunStackNotFound(t *testing.T) {
	yaml := `runtime: podman
stacks:
  - name: test
    path: /tmp
`
	path := writeTempFile(t, "deployer.yaml", yaml)
	code := run(path, "status", "nonexistent", mockRunner{})
	if code != 1 {
		t.Errorf("got exit code %d, want 1", code)
	}
}

func TestRunStatus(t *testing.T) {
	yaml := `runtime: podman
stacks:
  - name: test
    path: /tmp
`
	path := writeTempFile(t, "deployer.yaml", yaml)
	mock := mockRunner{output: `[{"Name":"test","State":"running"}]`}
	code := run(path, "status", "", mock)
	if code != 0 {
		t.Errorf("got exit code %d, want 0", code)
	}
}
