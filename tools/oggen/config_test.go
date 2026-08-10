package main

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "og.config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const validConfig = `{
  "projects": [
    {
      "name": "karako",
      "type": "Software & IT Services",
      "tagline": "Digital Intelligence.\nReal Impact.",
      "description": "Software engineering.",
      "url": "https://karakosystems.com",
      "background": "#2D2B55",
      "accent": ["#6C5CE7", "#8577ed"],
      "logoFull": "assets/full.png",
      "logoIcon": "assets/icon.png"
    }
  ]
}`

func TestLoadConfigValid(t *testing.T) {
	path := writeConfig(t, t.TempDir(), validConfig)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("projects = %d, want 1", len(cfg.Projects))
	}
	p := cfg.Projects[0]
	if p.Name != "karako" || p.Tagline != "Digital Intelligence.\nReal Impact." {
		t.Errorf("unexpected project: %+v", p)
	}
	if len(p.Accent) != 2 {
		t.Errorf("accent = %v", p.Accent)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	cases := map[string]struct {
		mutate string
		want   string
	}{
		"missing name":     {`"name": "karako"`, "name is required"},
		"missing tagline":  {`"tagline": "Digital Intelligence.\nReal Impact."`, "tagline is required"},
		"missing url":      {`"url": "https://karakosystems.com"`, "url is required"},
		"missing logoFull": {`"logoFull": "assets/full.png"`, "logoFull is required"},
		"missing logoIcon": {`"logoIcon": "assets/icon.png"`, "logoIcon is required"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			broken := strings.Replace(validConfig, tc.mutate, `"unused": "x"`, 1)
			path := writeConfig(t, t.TempDir(), broken)
			_, err := LoadConfig(path)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}

	t.Run("empty accent", func(t *testing.T) {
		broken := strings.Replace(validConfig, `["#6C5CE7", "#8577ed"]`, `[]`, 1)
		path := writeConfig(t, t.TempDir(), broken)
		if _, err := LoadConfig(path); err == nil || !strings.Contains(err.Error(), "accent") {
			t.Fatalf("err = %v, want accent error", err)
		}
	})

	t.Run("bad background", func(t *testing.T) {
		broken := strings.Replace(validConfig, "#2D2B55", "2D2B55", 1)
		path := writeConfig(t, t.TempDir(), broken)
		if _, err := LoadConfig(path); err == nil || !strings.Contains(err.Error(), "invalid color") {
			t.Fatalf("err = %v, want invalid color", err)
		}
	})

	t.Run("no projects", func(t *testing.T) {
		path := writeConfig(t, t.TempDir(), `{"projects": []}`)
		if _, err := LoadConfig(path); err == nil {
			t.Fatal("want error for empty projects")
		}
	})
}

func TestParseHexColor(t *testing.T) {
	got, err := parseHexColor("#2D2B55")
	if err != nil {
		t.Fatal(err)
	}
	if got != (color.NRGBA{R: 45, G: 43, B: 85, A: 255}) {
		t.Errorf("got %+v", got)
	}
	for _, bad := range []string{"2D2B55", "#2D2B5", "#GGGGGG", ""} {
		if _, err := parseHexColor(bad); err == nil {
			t.Errorf("parseHexColor(%q): want error", bad)
		}
	}
}

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, validConfig)
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := findRoot(nested)
	if err != nil {
		t.Fatalf("findRoot: %v", err)
	}
	rootEval, _ := filepath.EvalSymlinks(root)
	gotEval, _ := filepath.EvalSymlinks(got)
	if gotEval != rootEval {
		t.Errorf("findRoot = %s, want %s", got, root)
	}

	if _, err := findRoot(t.TempDir()); err == nil {
		t.Error("want error when og.config.json is absent")
	}
}
