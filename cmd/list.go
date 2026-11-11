package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/szyhf/go-docker-use/internal/dockerapi"
	"github.com/szyhf/go-docker-use/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有账号（* 标记各 registry 当前登录的账号）",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList() error {
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

	type row struct {
		alias string
		acct  store.Account
	}
	rows := make([]row, 0, len(ledger.Accounts))
	for alias, acct := range ledger.Accounts {
		rows = append(rows, row{alias, acct})
	}
	// 同 registry 的账号聚在一起，registry 内按别名排序。
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].acct.Registry != rows[j].acct.Registry {
			return rows[i].acct.Registry < rows[j].acct.Registry
		}
		return rows[i].alias < rows[j].alias
	})

	width := 0
	for _, r := range rows {
		if len(r.alias) > width {
			width = len(r.alias)
		}
	}
	// 每个 registry 查询一次当前登录用户（含 credential helper 调用），
	// 按 registry 缓存避免重复查询。
	currentByReg := map[string]string{}
	for _, r := range rows {
		if _, ok := currentByReg[r.acct.Registry]; !ok {
			currentByReg[r.acct.Registry] = dockerapi.CurrentUser(r.acct.Registry)
		}
	}

	for _, r := range rows {
		mark := " "
		if current := currentByReg[r.acct.Registry]; current != "" && current == r.acct.Username {
			mark = "*"
		}
		fmt.Printf("%s %-*s  %s@%s\n", mark, width, r.alias, r.acct.Username, dockerapi.DisplayName(r.acct.Registry))
	}
	return nil
}
