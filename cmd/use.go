package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/szyhf/go-docker-use/internal/dockerapi"
	"github.com/szyhf/go-docker-use/internal/store"
)

var useCmd = &cobra.Command{
	Use:   "use <别名>",
	Short: "切换到指定账号（等价于直接 docker-use <别名>）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUse(args[0])
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}

func runUse(alias string) error {
	st, err := store.New()
	if err != nil {
		return err
	}
	ledger, err := st.Load()
	if err != nil {
		return err
	}
	acct, ok := ledger.Accounts[alias]
	if !ok {
		return fmt.Errorf("没有别名为 %q 的账号，执行 docker-use list 查看全部", alias)
	}

	pw, err := st.GetPassword(alias)
	if err != nil {
		return err
	}

	// 先登出清理 config.json/钥匙串中的旧凭证，再登录写入新账号。
	dockerapi.Logout(acct.Registry)
	if err := dockerapi.Login(acct.Registry, acct.Username, []byte(pw)); err != nil {
		return fmt.Errorf("%w\n若密码已变更，请重新录入: docker-use add %s -u %s -r %s",
			err, alias, acct.Username, acct.Registry)
	}
	fmt.Printf("✓ 已切换到 %s (%s@%s)\n", alias, acct.Username, dockerapi.DisplayName(acct.Registry))
	return nil
}
