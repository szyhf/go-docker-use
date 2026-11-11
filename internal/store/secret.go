package store

import (
	"errors"
	"fmt"
	"os"

	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/zalando/go-keyring"
)

// keyringService 是密码在系统钥匙串中的服务名，条目 key 为账号别名。
const keyringService = "docker-use"

// ErrPasswordNotFound 表示别名的密码在钥匙串与降级文件中均不存在。
var ErrPasswordNotFound = errors.New("密码未找到，请重新执行 docker-use add 录入")

// SetPassword 保存别名对应的密码：优先写系统钥匙串；
// 钥匙串不可用时降级为本地明文文件（0600）。返回值指示是否走了钥匙串。
func (s *Store) SetPassword(alias, password string) (usedKeyring bool, err error) {
	if err := keyring.Set(keyringService, alias, password); err == nil {
		return true, nil
	}
	if err := s.saveFallback(alias, password); err != nil {
		return false, fmt.Errorf("保存密码失败: %w", err)
	}
	return false, nil
}

// GetPassword 读取别名对应的密码，先查钥匙串，再查降级文件。
func (s *Store) GetPassword(alias string) (string, error) {
	pw, keyringErr := keyring.Get(keyringService, alias)
	if keyringErr == nil {
		return pw, nil
	}
	if pw, ok := s.loadFallback()[alias]; ok {
		return pw, nil
	}
	if keyringErr != nil && !errors.Is(keyringErr, keyring.ErrNotFound) {
		return "", fmt.Errorf("读取钥匙串失败: %w; %w", keyringErr, ErrPasswordNotFound)
	}
	return "", ErrPasswordNotFound
}

// DeletePassword 删除别名的密码：钥匙串与降级文件都尽力清理。
func (s *Store) DeletePassword(alias string) error {
	_ = keyring.Delete(keyringService, alias) // 条目本就可能不存在，忽略
	fallback := s.loadFallback()
	if _, ok := fallback[alias]; !ok {
		return nil
	}
	delete(fallback, alias)
	return s.writeFallback(fallback)
}

// FallbackPath 返回降级密码文件路径（用于降级时的警告提示）。
func (s *Store) FallbackPath() string { return s.fallbackPath }

func (s *Store) loadFallback() map[string]string {
	data, err := os.ReadFile(s.fallbackPath)
	if err != nil {
		return nil
	}
	var m map[string]string
	if json.Unmarshal(data, &m) != nil {
		return nil
	}
	return m
}

func (s *Store) saveFallback(alias, password string) error {
	fallback := s.loadFallback()
	if fallback == nil {
		fallback = map[string]string{}
	}
	fallback[alias] = password
	return s.writeFallback(fallback)
}

func (s *Store) writeFallback(m map[string]string) error {
	data, err := json.Marshal(m, jsontext.Multiline(true), jsontext.SpaceAfterColon(true))
	if err != nil {
		return err
	}
	return os.WriteFile(s.fallbackPath, append(data, '\n'), 0o600)
}
