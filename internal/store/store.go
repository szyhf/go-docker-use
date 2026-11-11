// Package store 管理账号账本（accounts.json）的持久化。
// 账本只保存「别名 → registry + 用户名」映射，密码一律走 secret.go 的凭证存储。
package store

import (
	"fmt"
	"os"
	"path/filepath"

	"encoding/json/jsontext"
	"encoding/json/v2"
)

// Account 描述一个受管账号：全局唯一的别名映射到 registry 与用户名。
type Account struct {
	Registry string `json:"registry"`
	Username string `json:"username"`
}

// Ledger 是账本文件的完整内容。
type Ledger struct {
	Version  int                `json:"version"`
	Accounts map[string]Account `json:"accounts"`
}

const ledgerVersion = 1

// Store 提供账本与密码的读写，数据目录为 ~/.config/docker-use/。
type Store struct {
	dir          string
	ledgerPath   string
	fallbackPath string
}

// New 创建 Store 并确保数据目录存在（0700）。
func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法确定用户主目录: %w", err)
	}
	dir := filepath.Join(home, ".config", "docker-use")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("无法创建数据目录 %s: %w", dir, err)
	}
	return &Store{
		dir:          dir,
		ledgerPath:   filepath.Join(dir, "accounts.json"),
		fallbackPath: filepath.Join(dir, "passwords.json"),
	}, nil
}

// LedgerPath 返回账本文件路径。
func (s *Store) LedgerPath() string { return s.ledgerPath }

// Load 读取账本；文件不存在时返回空账本。
func (s *Store) Load() (*Ledger, error) {
	data, err := os.ReadFile(s.ledgerPath)
	if os.IsNotExist(err) {
		return &Ledger{Version: ledgerVersion, Accounts: map[string]Account{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("无法读取账本: %w", err)
	}
	var l Ledger
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("账本 %s 格式错误: %w", s.ledgerPath, err)
	}
	if l.Accounts == nil {
		l.Accounts = map[string]Account{}
	}
	return &l, nil
}

// Save 写回账本，权限 0600。
func (s *Store) Save(l *Ledger) error {
	l.Version = ledgerVersion
	data, err := json.Marshal(l, jsontext.Multiline(true), jsontext.SpaceAfterColon(true))
	if err != nil {
		return fmt.Errorf("序列化账本失败: %w", err)
	}
	if err := os.WriteFile(s.ledgerPath, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("无法写入账本: %w", err)
	}
	return nil
}
