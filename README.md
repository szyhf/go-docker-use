# docker-use

快速切换多个 Docker registry 账号的命令行工具。

为同一个 registry 下的多个账号（或多个 registry 的多个账号）分别建立别名，
一条命令完成切换，密码保存在系统钥匙串中，全程不回显、不进 shell 历史、
不出现在进程参数里。

## 安装

```bash
go install github.com/szyhf/go-docker-use@latest
```

本地构建：

```bash
go build -o docker-use .
```

## 使用

```bash
# 添加账号：交互输入密码（输入两次确认，不回显）
# -r 指定 registry，默认 Docker Hub
docker-use add work -u zhangsan -r harbor.l-feng.com
docker-use add hub-a -u alice

# 切换：一条命令（logout 旧凭证 + login 新账号）
docker-use work
# 等价写式
docker-use use work

# 查看状态
docker-use list      # 所有账号，* 标记各 registry 当前登录的账号
docker-use whoami    # 各 registry 的实时登录态（读取 docker 凭证体系）

# 删除账号（账本 + 钥匙串密码一并清理）
docker-use rm work          # 有确认
docker-use rm -y work       # 跳过确认
```

别名全局唯一（保留字如 `list`、`whoami`、`use` 等不可用作别名）。

## 工作原理

- **账本** `~/.config/docker-use/accounts.json`（0600）只保存
  「别名 → registry + 用户名」映射，不保存密码。
- **密码** 优先存系统钥匙串（macOS Keychain / Windows 凭据管理器 /
  Linux Secret Service，key 为 `docker-use/<别名>`）；钥匙串不可用时
  降级为 `~/.config/docker-use/passwords.json`（0600 明文）并在操作时警告。
- **切换** 通过 `docker login --password-stdin` 完成，密码经 stdin 管道
  传递，不出现在 argv（避免被 `ps` 捕获）；写入路径（auths / credsStore /
  credHelpers）全部由 docker 官方客户端处理。
- **whoami / list** 实时读取 `~/.docker/config.json` 并按需调用
  `docker-credential-*` helper 查询当前登录用户，再反查账本别名。

## 安全说明

- 密码输入使用 `term.ReadPassword`，终端不回显。
- 密码不会写入 shell 历史或进程参数。
- 降级明文文件的风险等级与 docker 默认的 `auths` 存储相同，仅在系统
  钥匙串不可用（如无 GUI 的 Linux）时才会启用，可随时删除该文件。
