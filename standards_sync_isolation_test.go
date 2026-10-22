package main

import (
	"os"
	"path/filepath"
	"testing"
)

// v2 必须使用独立的配置文件：旧版 App 留下的 standards-sync.json（指向旧 standards.git）
// 既不能覆盖 v2 的默认仓库，也不能被 v2 改写。
func TestStandardsSyncConfigIsolatedFromV1(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, "Library", "Application Support", "TeamStandards")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	oldFile := filepath.Join(dir, "standards-sync.json")
	oldBody := []byte(`{"repo_url":"https://github.com/kary2999/standards.git","branch":"main","last_synced_sha":"c7ed127"}`)
	if err := os.WriteFile(oldFile, oldBody, 0o600); err != nil {
		t.Fatal(err)
	}

	c, err := loadStandardsSyncConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.RepoURL != "" || c.LastSyncedSHA != "" {
		t.Fatalf("v2 must not inherit v1 state, got repo=%q sha=%q", c.RepoURL, c.LastSyncedSHA)
	}
	// 首次后台检查会把默认值（v2 仓库）写入配置；这里直接复现该步骤
	c.RepoURL, c.Branch = DefaultStandardsRepoURL, DefaultStandardsBranch
	if err := saveStandardsSyncConfig(c); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(oldFile); string(got) != string(oldBody) {
		t.Fatalf("v1 config file was modified: %s", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "standards-sync-v2.json")); err != nil {
		t.Fatalf("v2 config file missing: %v", err)
	}
	if DefaultStandardsRepoURL != "https://github.com/kary2999/standards-v2.git" {
		t.Fatalf("unexpected default repo %s", DefaultStandardsRepoURL)
	}
}
