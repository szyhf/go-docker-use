// Package dockerapi 封装对 docker 客户端凭证体系的只读查询与登录操作。
// 本包不直接写 ~/.docker/config.json —— 所有写入都通过 docker CLI 完成，
// 以便正确处理 credsStore / credHelpers 等凭证存储差异。
package dockerapi

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"encoding/json/v2"
)

type authEntry struct {
	Auth string `json:"auth"`
}

type config struct {
	CredsStore  string               `json:"credsStore"`
	CredHelpers map[string]string    `json:"credHelpers"`
	Auths       map[string]authEntry `json:"auths"`
}

func loadConfig() *config {
	home, err := os.UserHomeDir()
	if err != nil {
		return &config{}
	}
	data, err := os.ReadFile(filepath.Join(home, ".docker", "config.json"))
	if err != nil {
		return &config{}
	}
	var c config
	if json.Unmarshal(data, &c) != nil {
		return &config{}
	}
	return &c
}

// CurrentUser 返回指定 registry 当前在 docker 侧的登录用户名；
// 未登录或无法确定时返回空串。
// 查询顺序与 docker 客户端一致：credHelpers（按 registry）→ credsStore（全局）→ auths 明文。
func CurrentUser(registry string) string {
	cfg := loadConfig()
	keys := matchKeys(cfg, registry)

	for _, key := range keys {
		if helper, ok := cfg.CredHelpers[key]; ok {
			if u := helperUsername(helper, key); u != "" {
				return u
			}
		}
	}
	if cfg.CredsStore != "" {
		for _, key := range keys {
			if u := helperUsername(cfg.CredsStore, key); u != "" {
				return u
			}
		}
	}
	for _, key := range keys {
		if entry, ok := cfg.Auths[key]; ok && entry.Auth != "" {
			if raw, err := base64.StdEncoding.DecodeString(entry.Auth); err == nil {
				if u, _, ok := strings.Cut(string(raw), ":"); ok && u != "" {
					return u
				}
			}
		}
	}
	return ""
}

// matchKeys 在 config.json 的 auths/credHelpers 键中找出与 registry 对应的
// 原始键：先精确匹配，再按归一化形式模糊匹配
// （历史条目的键可能带或不带 https:// 前缀，会指向不同的钥匙串条目）。
func matchKeys(cfg *config, registry string) []string {
	var keys []string
	if _, ok := cfg.Auths[registry]; ok {
		keys = append(keys, registry)
	}
	if _, ok := cfg.CredHelpers[registry]; ok {
		keys = append(keys, registry)
	}
	target := NormalizeRegistry(registry)
	for key := range cfg.Auths {
		if key != registry && NormalizeRegistry(key) == target {
			keys = append(keys, key)
		}
	}
	for key := range cfg.CredHelpers {
		if key != registry && NormalizeRegistry(key) == target {
			keys = append(keys, key)
		}
	}
	return keys
}

// helperUsername 调用 docker-credential-<helper> get 查询钥匙串中的凭证，
// 未登录（helper 返回 not found）或 helper 执行失败时返回空串。
func helperUsername(helper, serverURL string) string {
	cmd := exec.Command("docker-credential-"+helper, "get")
	cmd.Stdin = strings.NewReader(serverURL)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	var cred struct {
		Username string `json:"Username"`
	}
	if json.Unmarshal(out, &cred) != nil {
		return ""
	}
	return cred.Username
}
