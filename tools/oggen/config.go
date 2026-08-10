package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
)

type Config struct {
	Projects []Project `json:"projects"`
}

type Texts struct {
	Type        string `json:"type"`
	Tagline     string `json:"tagline"`
	Description string `json:"description"`
}

type Project struct {
	Name          string           `json:"name"`
	Type          string           `json:"type"`
	Tagline       string           `json:"tagline"`
	Description   string           `json:"description"`
	URL           string           `json:"url"`
	Background    string           `json:"background"`
	Accent        []string         `json:"accent"`
	LogoFull      string           `json:"logoFull"`
	LogoIcon      string           `json:"logoIcon"`
	Locales       map[string]Texts `json:"locales"`
	DefaultLocale string           `json:"defaultLocale"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- dev-time CLI reading the repo-local config
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(cfg.Projects) == 0 {
		return nil, fmt.Errorf("%s: no projects defined", path)
	}
	for i, p := range cfg.Projects {
		if err := p.validate(); err != nil {
			return nil, fmt.Errorf("project %d (%q): %w", i, p.Name, err)
		}
	}
	return &cfg, nil
}

func (p Project) validate() error {
	switch {
	case p.Name == "":
		return fmt.Errorf("name is required")
	case p.URL == "":
		return fmt.Errorf("url is required")
	case p.LogoFull == "":
		return fmt.Errorf("logoFull is required")
	case p.LogoIcon == "":
		return fmt.Errorf("logoIcon is required")
	case len(p.Accent) == 0:
		return fmt.Errorf("accent needs at least one color")
	}
	if len(p.Locales) == 0 {
		if p.Tagline == "" {
			return fmt.Errorf("tagline is required")
		}
	} else {
		if p.Type != "" || p.Tagline != "" || p.Description != "" {
			return fmt.Errorf("top-level texts and locales are mutually exclusive")
		}
		if p.DefaultLocale == "" {
			return fmt.Errorf("defaultLocale is required when locales are set")
		}
		if _, ok := p.Locales[p.DefaultLocale]; !ok {
			return fmt.Errorf("defaultLocale %q is not defined in locales", p.DefaultLocale)
		}
		for lang, t := range p.Locales {
			if t.Tagline == "" {
				return fmt.Errorf("locale %q: tagline is required", lang)
			}
		}
	}
	if _, err := parseHexColor(p.Background); err != nil {
		return fmt.Errorf("background: %w", err)
	}
	for _, c := range p.Accent {
		if _, err := parseHexColor(c); err != nil {
			return fmt.Errorf("accent: %w", err)
		}
	}
	return nil
}

func parseHexColor(s string) (color.NRGBA, error) {
	if len(s) != 7 || s[0] != '#' {
		return color.NRGBA{}, fmt.Errorf("invalid color %q (expected #rrggbb)", s)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid color %q: %w", s, err)
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
}

func findRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "og.config.json")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("og.config.json not found from %s upward", start)
		}
		dir = parent
	}
}
