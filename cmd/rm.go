package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/szyhf/go-docker-use/internal/dockerapi"
	"github.com/szyhf/go-docker-use/internal/prompt"
	"github.com/szyhf/go-docker-use/internal/store"
)

var rmYes bool

var rmCmd = &cobra.Command{
	Use:   "rm <别名>",
	Short: "删除账号（账本与钥匙串中的密码一并清理）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRm(args[0])
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmYes, "yes", "y", false, "跳过删除确认")
	rootCmd.AddCommand(rmCmd)
}

func runRm(alias string) error {
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

	if !rmYes {
		ok, err := prompt.Confirm(
			fmt.Sprintf("删除账号 %s (%s@%s)？", alias, acct.Username, dockerapi.DisplayName(acct.Registry)),
			false)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "已取消。")
			return nil
		}
	}

	delete(ledger.Accounts, alias)
	if err := st.Save(ledger); err != nil {
		return err
	}
	if err := st.DeletePassword(alias); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ 账本已更新，但清理密码失败: %v\n", err)
	}
	fmt.Printf("✓ 已删除 %s\n", alias)
	return nil
}
