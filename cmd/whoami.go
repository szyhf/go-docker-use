package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/szyhf/go-docker-use/internal/dockerapi"
	"github.com/szyhf/go-docker-use/internal/store"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "显示各 registry 的当前登录账号（实时读取 docker 登录态）",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWhoami()
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami() error {
	st, err := store.New()
	if err != nil {
		return err
	}
	ledger, err := st.Load()
	if err != nil {
		return err
	}
	if len(ledger.Accounts) == 0 {
		fmt.Println("还没有任何账号，使用 docker-use add <别名> -u <用户名> 添加")
		return nil
	}

	// 账本中出现过的 registry 去重排序，逐个查询当前登录用户并反查别名。
	registries := map[string]bool{}
	for _, acct := range ledger.Accounts {
		registries[acct.Registry] = true
	}
	keys := make([]string, 0, len(registries))
	for reg := range registries {
		keys = append(keys, reg)
	}
	sort.Strings(keys)

	width := 0
	for _, reg := range keys {
		if name := dockerapi.DisplayName(reg); len(name) > width {
			width = len(name)
		}
	}

	for _, reg := range keys {
		user := dockerapi.CurrentUser(reg)
		if user == "" {
			fmt.Printf("%-*s  未登录\n", width, dockerapi.DisplayName(reg))
			continue
		}
		alias := reverseLookup(ledger, reg, user)
		if alias == "" {
			// docker 侧有登录但不在账本内（手动 docker login 过），原样显示。
			fmt.Printf("%-*s  %s（不在账本中）\n", width, dockerapi.DisplayName(reg), user)
			continue
		}
		fmt.Printf("%-*s  %s (%s)\n", width, dockerapi.DisplayName(reg), user, alias)
	}
	return nil
}

// reverseLookup 在账本中查找 registry+用户名 对应的别名。
func reverseLookup(ledger *store.Ledger, registry, username string) string {
	for alias, acct := range ledger.Accounts {
		if acct.Registry == registry && acct.Username == username {
			return alias
		}
	}
	return ""
}
