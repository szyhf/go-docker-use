// Package cmd 定义 docker-use 的所有命令。
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/szyhf/go-docker-use/internal/dockerapi"
)

var rootCmd = &cobra.Command{
	Use:   "docker-use",
	Short: "快速切换多个 Docker registry 账号",
	Long: `docker-use 管理多个 Docker registry 账号，一条命令完成切换。

密码保存在系统钥匙串中（不可用时降级为 0600 本地文件并警告），
切换通过 docker login --password-stdin 完成，密码不会出现在进程参数里。`,
	Args: cobra.ArbitraryArgs,
	// SilenceUsage: 业务错误（别名不存在、登录失败等）不需要打印 usage。
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// 第一个参数不是已注册子命令时，按账号别名处理，直接切换。
		return runUse(args[0])
	},
}

// Execute 是命令行入口。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// reservedNames 返回禁止用作账号别名的保留字：
// 全部注册子命令名加上常见缩写，防止别名劫持子命令路由。
func reservedNames() map[string]bool {
	res := map[string]bool{
		"help": true, "completion": true, "use": true,
		"ls": true, "remove": true, "delete": true, "del": true, "rm": true,
	}
	for _, c := range rootCmd.Commands() {
		res[c.Name()] = true
	}
	return res
}

// hubRegistryFlag 是 add 命令 -r 的默认值。
var hubRegistryFlag = dockerapi.DockerHubKey
