package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/szyhf/go-docker-use/internal/dockerapi"
	"github.com/szyhf/go-docker-use/internal/prompt"
	"github.com/szyhf/go-docker-use/internal/store"
)

var addUsername string
var addRegistry string

var addCmd = &cobra.Command{
	Use:   "add <别名> -u <用户名> [-r <registry>]",
	Short: "添加账号（交互式输入密码，不回显）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(args[0])
	},
}

func init() {
	addCmd.Flags().StringVarP(&addUsername, "username", "u", "", "账号用户名（必填）")
	addCmd.Flags().StringVarP(&addRegistry, "registry", "r", hubRegistryFlag,
		"registry 地址，默认 Docker Hub")
	_ = addCmd.MarkFlagRequired("username")
	rootCmd.AddCommand(addCmd)
}

func runAdd(alias string) error {
	if reservedNames()[alias] {
		return fmt.Errorf("别名 %q 是保留字，请换一个", alias)
	}
	if strings.ContainsAny(alias, " \t/\\") {
		return fmt.Errorf("别名不能包含空白或斜杠字符")
	}

	st, err := store.New()
	if err != nil {
		return err
	}
	ledger, err := st.Load()
	if err != nil {
		return err
	}

	if old, ok := ledger.Accounts[alias]; ok {
		ok, err := prompt.Confirm(
			fmt.Sprintf("别名 %s 已存在（%s@%s），覆盖？", alias, old.Username, dockerapi.DisplayName(old.Registry)),
			false)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "已取消。")
			return nil
		}
	}

	registryKey := dockerapi.CanonicalKey(addRegistry)
	pw, err := prompt.ReadPasswordTwice(
		fmt.Sprintf("输入 %s@%s 的密码: ", addUsername, dockerapi.DisplayName(registryKey)))
	if err != nil {
		return err
	}

	usedKeyring, err := st.SetPassword(alias, pw)
	if err != nil {
		return err
	}
	ledger.Accounts[alias] = store.Account{Registry: registryKey, Username: addUsername}
	if err := st.Save(ledger); err != nil {
		return err
	}

	fmt.Printf("✓ 已添加 %s (%s@%s)\n", alias, addUsername, dockerapi.DisplayName(registryKey))
	if !usedKeyring {
		fmt.Fprintf(os.Stderr, "⚠ 系统钥匙串不可用，密码已降级保存到 %s（0600 明文）\n", st.FallbackPath())
	}
	fmt.Printf("切换账号：docker-use %s\n", alias)
	return nil
}
