package main

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: write a temp file with given content, return its path
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantKeys   int
		wantErrors int
	}{
		{
			name:     "valid file",
			content:  "DB_HOST=localhost\nDB_PORT=5432\n",
			wantKeys: 2,
		},
		{
			name:     "skips comments and empty lines",
			content:  "# comment\n\nDB_HOST=localhost\n\n# another\nDB_PORT=5432\n",
			wantKeys: 2,
		},
		{
			name:       "invalid syntax",
			content:    "DB_HOST=localhost\nINVALID LINE\nDB_PORT=5432\n",
			wantKeys:   2,
			wantErrors: 1,
		},
		{
			name:     "strips quotes",
			content:  "SECRET=\"mypassword\"\nTOKEN='abc123'\n",
			wantKeys: 2,
		},
		{
			name:     "empty value",
			content:  "DB_HOST=\nDB_PORT=5432\n",
			wantKeys: 2,
		},
		{
			name:     "spaces around equals",
			content:  "DB_HOST = localhost\n",
			wantKeys: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, ".env", tt.content)
			entries, problems := parseEnvFile(path)

			if len(entries) != tt.wantKeys {
				t.Errorf("got %d entries, want %d", len(entries), tt.wantKeys)
			}
			if len(problems) != tt.wantErrors {
				t.Errorf("got %d parse errors, want %d", len(problems), tt.wantErrors)
			}
		})
	}
}

func TestParseEnvFileNotFound(t *testing.T) {
	entries, _ := parseEnvFile("/nonexistent/.env")
	if entries != nil {
		t.Error("expected nil entries for missing file")
	}
}

func TestFindDuplicates(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		want    int
	}{
		{
			name: "no duplicates",
			entries: []Entry{
				{Key: "A", Line: 1},
				{Key: "B", Line: 2},
			},
			want: 0,
		},
		{
			name: "one duplicate",
			entries: []Entry{
				{Key: "A", Line: 1},
				{Key: "B", Line: 2},
				{Key: "A", Line: 3},
			},
			want: 1,
		},
		{
			name: "triple duplicate",
			entries: []Entry{
				{Key: "A", Line: 1},
				{Key: "A", Line: 2},
				{Key: "A", Line: 3},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := findDuplicates(tt.entries)
			if len(problems) != tt.want {
				t.Errorf("got %d duplicates, want %d", len(problems), tt.want)
			}
			for _, p := range problems {
				if p.Level != "error" {
					t.Errorf("duplicate should be error, got %s", p.Level)
				}
			}
		})
	}
}

func TestFindEmptyValues(t *testing.T) {
	entries := []Entry{
		{Key: "A", Value: "hello", Line: 1},
		{Key: "B", Value: "", Line: 2},
		{Key: "C", Value: "world", Line: 3},
		{Key: "D", Value: "", Line: 4},
	}

	problems := findEmptyValues(entries)
	if len(problems) != 2 {
		t.Errorf("got %d empty warnings, want 2", len(problems))
	}
	for _, p := range problems {
		if p.Level != "warn" {
			t.Errorf("empty value should be warn, got %s", p.Level)
		}
	}
}

func TestCompareWithExample(t *testing.T) {
	env := []Entry{
		{Key: "A", Value: "1", Line: 1},
		{Key: "B", Value: "2", Line: 2},
		{Key: "EXTRA", Value: "3", Line: 3},
	}
	example := []Entry{
		{Key: "A", Value: "", Line: 1},
		{Key: "B", Value: "", Line: 2},
		{Key: "MISSING", Value: "", Line: 3},
	}

	problems := compareWithExample(env, example, ".env", ".env.example")

	var missing, extra int
	for _, p := range problems {
		if p.Level == "error" {
			missing++
		} else {
			extra++
		}
	}

	if missing != 1 {
		t.Errorf("got %d missing keys, want 1", missing)
	}
	if extra != 1 {
		t.Errorf("got %d extra keys, want 1", extra)
	}
}

func TestRunExitCodes(t *testing.T) {
	// Disable colors for test output
	cReset, cOlive, cRust, cPurple, cMuted, cBold = "", "", "", "", "", ""

	t.Run("file not found returns 2", func(t *testing.T) {
		code := run("/nonexistent/.env", "/nonexistent/.env.example", true)
		if code != 2 {
			t.Errorf("got exit code %d, want 2", code)
		}
	})

	t.Run("clean file returns 0", func(t *testing.T) {
		env := writeTempFile(t, ".env", "A=hello\nB=world\n")
		example := writeTempFile(t, ".env.example", "A=\nB=\n")
		code := run(env, example, true)
		if code != 0 {
			t.Errorf("got exit code %d, want 0", code)
		}
	})

	t.Run("errors return 1", func(t *testing.T) {
		env := writeTempFile(t, ".env", "A=hello\n")
		example := writeTempFile(t, ".env.example", "A=\nMISSING=\n")
		code := run(env, example, true)
		if code != 1 {
			t.Errorf("got exit code %d, want 1", code)
		}
	})

	t.Run("warnings only return 0", func(t *testing.T) {
		env := writeTempFile(t, ".env", "A=hello\nEXTRA=something\n")
		example := writeTempFile(t, ".env.example", "A=\n")
		code := run(env, example, true)
		if code != 0 {
			t.Errorf("got exit code %d, want 0 (warnings are not errors)", code)
		}
	})
}
