// Package prompt 处理交互式安全输入：密码读取不回显，确认问答支持默认值。
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var stdinReader = bufio.NewReader(os.Stdin)

// ReadPassword 提示并读取密码：不回显、不进入 shell 历史，提示打到 stderr。
func ReadPassword(hint string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s", hint)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
	}
	return string(b), nil
}

// ReadPasswordTwice 读取两次密码并确认一致，最多重试 3 次。
func ReadPasswordTwice(hint string) (string, error) {
	const maxAttempts = 3
	for i := 0; i < maxAttempts; i++ {
		pw, err := ReadPassword(hint)
		if err != nil {
			return "", err
		}
		pw2, err := ReadPassword("再次输入以确认: ")
		if err != nil {
			return "", err
		}
		if pw == pw2 {
			return pw, nil
		}
		fmt.Fprintln(os.Stderr, "两次输入不一致，请重试。")
	}
	return "", fmt.Errorf("密码确认失败次数过多")
}

// Confirm 输出问题并等待 y/n 确认，直接回车采用默认值；
// stdin 非 TTY（脚本/管道）时直接返回默认值。
func Confirm(question string, def bool) (bool, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return def, nil
	}
	suffix := " [y/N]: "
	if def {
		suffix = " [Y/n]: "
	}
	fmt.Fprintf(os.Stderr, "%s%s", question, suffix)
	line, err := stdinReader.ReadString('\n')
	if err != nil && line == "" {
		// EOF（stdin 关闭或输入被截断）按默认值处理，不视为错误。
		return def, nil
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "":
		return def, nil
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
