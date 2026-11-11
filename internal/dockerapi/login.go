package dockerapi

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Login 执行 docker login。密码经 stdin 传递，不会出现在进程参数中
// （避免被 ps 或 shell 历史捕获）。Docker Hub 登录不带 registry 参数。
func Login(registry, username string, password []byte) error {
	args := []string{"login", "-u", username, "--password-stdin"}
	if !IsDockerHub(registry) {
		args = append(args, registry)
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdin = bytes.NewReader(password)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		detail := bytes.TrimSpace(out.Bytes())
		if len(detail) > 0 {
			return fmt.Errorf("docker login 失败: %s", detail)
		}
		return fmt.Errorf("docker login 失败: %w", err)
	}
	return nil
}

// Logout 清除 registry 的现有凭证，避免 config.json/钥匙串中的旧条目残留。
// 该操作幂等且失败不视为错误（未登录时 docker logout 本身也返回成功）。
func Logout(registry string) {
	args := []string{"logout"}
	if !IsDockerHub(registry) {
		args = append(args, registry)
	}
	_ = exec.Command("docker", args...).Run()
}
