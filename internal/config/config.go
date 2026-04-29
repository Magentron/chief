package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const configFile = ".chief/config.yaml"

// Config holds project-level settings for Chief.
type Config struct {
	Worktree   WorktreeConfig   `yaml:"worktree"`
	OnComplete OnCompleteConfig `yaml:"onComplete"`
	Agent      AgentConfig      `yaml:"agent"`
	Bash       BashConfig       `yaml:"bash"`
}

// BashConfig holds settings for external bash commands invoked by Chief
// (currently the worktree setup command).
type BashConfig struct {
	// Timeout is a Go duration string (e.g. "30s", "5m"). Empty values use
	// DefaultBashTimeout. Unparseable or negative values fall back to the
	// default and surface a warning via Config.BashTimeoutWarning.
	Timeout string `yaml:"timeout"`
}

// DefaultBashTimeout is applied when bash.timeout is unset or unparseable.
// Setup commands rarely need longer than this; users with slow installers
// should configure an explicit value.
const DefaultBashTimeout = 5 * time.Minute

// BashTimeout returns the configured bash command timeout as a time.Duration.
// Empty values use DefaultBashTimeout; unparseable or negative values also
// fall back to the default (BashTimeoutWarning describes the fallback).
// An explicit "0s" returns 0, which callers interpret as "no timeout".
// Surrounding whitespace in the configured value is ignored.
//
// Nil-safe: returns DefaultBashTimeout when c is nil so callers do not have to
// guard a missing config.
func (c *Config) BashTimeout() time.Duration {
	if c == nil {
		return DefaultBashTimeout
	}
	v := strings.TrimSpace(c.Bash.Timeout)
	if v == "" {
		return DefaultBashTimeout
	}
	d, err := time.ParseDuration(v)
	if err != nil || d < 0 {
		return DefaultBashTimeout
	}
	return d
}

// BashTimeoutWarning returns a human-readable warning when the configured
// bash.timeout value is non-empty but unparseable or negative, in which case
// BashTimeout silently falls back to DefaultBashTimeout. Returns "" when the
// value is empty (default), valid, or when c is nil.
func (c *Config) BashTimeoutWarning() string {
	if c == nil {
		return ""
	}
	v := strings.TrimSpace(c.Bash.Timeout)
	if v == "" {
		return ""
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Sprintf("bash.timeout %q is not a valid duration; using default %s", v, DefaultBashTimeout)
	}
	if d < 0 {
		return fmt.Sprintf("bash.timeout %q is negative; using default %s", v, DefaultBashTimeout)
	}
	return ""
}

// AgentConfig holds agent CLI settings (Claude, Codex, OpenCode, or Cursor).
type AgentConfig struct {
	Provider string `yaml:"provider"` // "claude" (default) | "codex" | "opencode" | "cursor"
	CLIPath  string `yaml:"cliPath"`  // optional custom path to CLI binary
}

// WorktreeConfig holds worktree-related settings.
type WorktreeConfig struct {
	Setup string `yaml:"setup"`
}

// OnCompleteConfig holds post-completion automation settings.
type OnCompleteConfig struct {
	Push     bool `yaml:"push"`
	CreatePR bool `yaml:"createPR"`
}

// Default returns a Config with zero-value defaults.
func Default() *Config {
	return &Config{}
}

// configPath returns the full path to the config file.
func configPath(baseDir string) string {
	return filepath.Join(baseDir, configFile)
}

// Exists checks if the config file exists.
func Exists(baseDir string) bool {
	_, err := os.Stat(configPath(baseDir))
	return err == nil
}

// Load reads the config from .chief/config.yaml.
// Returns Default() when the file doesn't exist (no error).
func Load(baseDir string) (*Config, error) {
	path := configPath(baseDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the config to .chief/config.yaml.
func Save(baseDir string, cfg *Config) error {
	path := configPath(baseDir)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
