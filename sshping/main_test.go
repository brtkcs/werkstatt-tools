package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	cOlive, cRust, cMuted, cReset = "", "", "", ""
}

// mockDialer simulates TCP connections for testing
type mockDialer struct {
	up  bool
	dur time.Duration
}

func (m mockDialer) dial(address string, timeout time.Duration) (bool, time.Duration) {
	return m.up, m.dur
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
	yaml := `hosts:
  - name: server1
    address: 10.0.0.1
    port: 22
  - name: server2
    address: 10.0.0.2
    port: 8080
`
	path := writeTempFile(t, "hosts.yaml", yaml)
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Hosts) != 2 {
		t.Fatalf("got %d hosts, want 2", len(cfg.Hosts))
	}
	if cfg.Hosts[0].Name != "server1" {
		t.Errorf("got name %q, want server1", cfg.Hosts[0].Name)
	}
	if cfg.Hosts[1].Port != 8080 {
		t.Errorf("got port %d, want 8080", cfg.Hosts[1].Port)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := loadConfig("/nonexistent/hosts.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	path := writeTempFile(t, "hosts.yaml", "{{invalid")
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
			cfg: Config{Hosts: []Host{
				{Name: "srv", Address: "10.0.0.1", Port: 22},
			}},
		},
		{
			name:    "no hosts",
			cfg:     Config{Hosts: []Host{}},
			wantErr: true,
		},
		{
			name: "missing name",
			cfg: Config{Hosts: []Host{
				{Name: "", Address: "10.0.0.1", Port: 22},
			}},
			wantErr: true,
		},
		{
			name: "missing address",
			cfg: Config{Hosts: []Host{
				{Name: "srv", Address: "", Port: 22},
			}},
			wantErr: true,
		},
		{
			name: "invalid port zero",
			cfg: Config{Hosts: []Host{
				{Name: "srv", Address: "10.0.0.1", Port: 0},
			}},
			wantErr: true,
		},
		{
			name: "invalid port too high",
			cfg: Config{Hosts: []Host{
				{Name: "srv", Address: "10.0.0.1", Port: 70000},
			}},
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

func TestPingHostsAllUp(t *testing.T) {
	hosts := []Host{
		{Name: "srv1", Address: "10.0.0.1", Port: 22},
		{Name: "srv2", Address: "10.0.0.2", Port: 22},
	}

	mock := mockDialer{up: true, dur: 5 * time.Millisecond}
	results := pingHosts(hosts, 2*time.Second, mock)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if !r.Up {
			t.Errorf("expected %s to be up", r.Host.Name)
		}
	}
}

func TestPingHostsAllDown(t *testing.T) {
	hosts := []Host{
		{Name: "srv1", Address: "10.0.0.1", Port: 22},
	}

	mock := mockDialer{up: false, dur: 2 * time.Second}
	results := pingHosts(hosts, 2*time.Second, mock)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Up {
		t.Error("expected host to be down")
	}
}

func TestRunAllUp(t *testing.T) {
	yaml := `hosts:
  - name: server
    address: 10.0.0.1
    port: 22
`
	path := writeTempFile(t, "hosts.yaml", yaml)
	mock := mockDialer{up: true, dur: 5 * time.Millisecond}
	code := run(path, 2*time.Second, mock)
	if code != 0 {
		t.Errorf("got exit code %d, want 0", code)
	}
}

func TestRunSomeDown(t *testing.T) {
	yaml := `hosts:
  - name: server
    address: 10.0.0.1
    port: 22
`
	path := writeTempFile(t, "hosts.yaml", yaml)
	mock := mockDialer{up: false, dur: 2 * time.Second}
	code := run(path, 2*time.Second, mock)
	if code != 1 {
		t.Errorf("got exit code %d, want 1 (some hosts down)", code)
	}
}

func TestRunMissingConfig(t *testing.T) {
	mock := mockDialer{up: true, dur: 0}
	code := run("/nonexistent/hosts.yaml", 2*time.Second, mock)
	if code != 2 {
		t.Errorf("got exit code %d, want 2", code)
	}
}
