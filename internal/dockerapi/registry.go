package dockerapi

import "strings"

// DockerHubKey 是 Docker Hub 凭证在 ~/.docker/config.json 中的规范键。
const DockerHubKey = "https://index.docker.io/v1/"

// NormalizeRegistry 把用户输入或 config.json 中的 registry 键归一化为
// 可比较的形式：小写 host[:port]，去掉 scheme 与路径；
// Docker Hub 的各种写法统一归一为 "index.docker.io"。
func NormalizeRegistry(registry string) string {
	r := strings.TrimSpace(strings.ToLower(registry))
	if i := strings.Index(r, "://"); i >= 0 {
		r = r[i+3:]
	}
	if i := strings.Index(r, "/"); i >= 0 {
		r = r[:i]
	}
	switch r {
	case "", "docker.io", "index.docker.io", "hub.docker.com",
		"registry.hub.docker.com", "registry-1.docker.io", "hub":
		return "index.docker.io"
	}
	return r
}

// IsDockerHub 判断 registry 键是否指向 Docker Hub。
func IsDockerHub(registry string) bool {
	return NormalizeRegistry(registry) == "index.docker.io"
}

// CanonicalKey 把用户输入（可能带 scheme、是 Docker Hub 变体）规范为
// 账本与 config.json 使用的键：Docker Hub 存规范键，私有 registry 存 host[:port]。
func CanonicalKey(registry string) string {
	if IsDockerHub(registry) {
		return DockerHubKey
	}
	return NormalizeRegistry(registry)
}

// DisplayName 返回 registry 的显示名，Docker Hub 特殊显示。
func DisplayName(registry string) string {
	if IsDockerHub(registry) {
		return "Docker Hub"
	}
	return registry
}
