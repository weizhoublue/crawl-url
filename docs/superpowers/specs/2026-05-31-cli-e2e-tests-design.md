# crawl-url CLI 端到端 (E2E) 测试设计规约

本文档定义了 `crawl-url` CLI 工具的端到端 (E2E) 测试用例及执行框架。通过在本地动态启动 HTTP Mock 服务器并直接调用编译好的 CLI 二进制文件，验证命令行参数处理、抓取限制（层级深度与数量）、文件保存和容错表现是否符合预期。

## 1. 测试框架设计

测试用例采用 Go 原生的 `go test` 工具链实施，所有测试文件均存放于新创建的 [tests/](file:///Users/weizhoublue/Documents/git/crawl-url/tests) 目录中。

```
tests/
├── main_test.go             # 全局构建脚手架与通用辅助逻辑
├── crawl_basic_test.go      # 基础递归抓取测试
├── crawl_limit_test.go      # --url-limit 数量限制测试
├── crawl_depth_test.go      # --depth 深度限制测试
├── crawl_exclude_test.go    # --exclude-prefix 排除前缀测试
├── crawl_media_test.go      # --image / --video 媒体流抓取测试
├── crawl_output_test.go     # --output-dir 页面保存测试
├── crawl_cli_error_test.go  # CLI 参数格式与错误验证（反向）
└── crawl_http_error_test.go # HTTP 404/500 等服务器异常处理（反向）
```

### 1.1 全局构建机制 (TestMain)
为避免每个测试用例重复编译带来的时间开销，在 [tests/main_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/main_test.go) 中使用 `TestMain` 进行一次性编译：
1. 在系统临时目录下生成唯一的临时执行文件路径。
2. 运行 `go build -o <tmp_dir>/crawl-url ../cmd/crawl-url/main.go ...` 编译最新的二进制文件。
3. 运行 `m.Run()` 执行所有测试用例。
4. 测试结束后，自动清理临时目录。

### 1.2 CLI 进程执行辅助函数 (runCLI)
为各用例提供标准的进程启动与输出捕获接口：
```go
func runCLI(t *testing.T, args []string) (stdout string, stderr string, exitCode int, err error)
```
该函数负责启动临时编译的 CLI，捕获输出，并处理 `os/exec` 的退出状态码（尤其针对退出码为 1 或 2 的异常用例）。

---

## 2. 测试用例设计

### 2.1 正向用例 (Positive Cases)

| 测试名称 | 测试文件 | 拓扑设计 / 种子内容 | 运行命令 | 预期结果与校验逻辑 |
| :--- | :--- | :--- | :--- | :--- |
| **基础递归抓取** | [crawl_basic_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_basic_test.go) | `/` -> 包含 `/page1`<br>`/page1` -> 包含 `/page2` 和外部链接 `https://other.com` | `crawl-url <server_url>` | 1. 退出码为 `0`<br>2. 输出精确包含 `/`, `/page1`, `/page2` 的全路径<br>3. 输出不包含 `https://other.com` |
| **URL 数量限制** | [crawl_limit_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_limit_test.go) | `/` -> 包含链接到 `/a`, `/b`, `/c`, `/d` | `crawl-url <server_url> --url-limit 2` | 1. 退出码为 `0`<br>2. 爬取出的 URL 数量精确等于 `2` (种子 `/` 占 1 个，子页面占 1 个) |
| **层级深度限制** | [crawl_depth_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_depth_test.go) | `/` (D0) -> `/d1` (D1)<br>`/d1` -> `/d2` (D2) | `crawl-url <server_url> --depth 2` | 1. 退出码为 `0`<br>2. 输出包含 `/` 和 `/d1`<br>3. **层级校验**：输出中绝对不包含 `/d2`（最后一层已停止向下扩展） |
| **排除前缀** | [crawl_exclude_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_exclude_test.go) | `/` -> 包含 `/allowed` 和 `/static/img.png` | `crawl-url <server_url> --exclude-prefix <server_url>/static/` | 1. 退出码为 `0`<br>2. 输出包含 `/allowed`，但绝对不含 `/static/img.png` |
| **媒体元素提取** | [crawl_media_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_media_test.go) | `/` -> 包含 `<img src="/pic.png">`<br>和 `<video src="/mov.mp4">` | ① `crawl-url <server_url>` <br>② `crawl-url <server_url> --image --video`<br>③ `crawl-url <server_url> --vedio` | 场景①：输出不含 `/pic.png` 和 `/mov.mp4`<br>场景②：输出包含 `/pic.png` 和 `/mov.mp4`<br>场景③：输出仅包含 `/mov.mp4` |
| **本地保存 HTML** | [crawl_output_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_output_test.go) | `/` -> 包含 `/page1` | `crawl-url <server_url> --output-dir <temp_dir>` | 1. 退出码为 `0`<br>2. 检测 `<temp_dir>` 目录下正确生成了两个 HTML 文件，且内容符合预期 |

### 2.2 反向用例 (Negative Cases)

#### CLI 输入校验异常 ([crawl_cli_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_cli_error_test.go))
1. **缺少种子 URL**：
   * 命令：`crawl-url`
   * 预期：退出码为 `2`，`stderr` 包含 `缺少种子 URL`。
2. **不支持短参数**：
   * 命令：`crawl-url <server_url> -workers 4`
   * 预期：退出码为 `2`，`stderr` 包含 `请使用 -- 长选项`。
3. **未知参数**：
   * 命令：`crawl-url <server_url> --foo`
   * 预期：退出码为 `2`，`stderr` 包含 `未知选项: --foo`。
4. **无效的并发 worker 数**：
   * 命令：`crawl-url <server_url> --workers 0` 或 `crawl-url <server_url> --workers -1`
   * 预期：退出码为 `1`，`stderr` 包含 `workers` 校验提示（如需大于 0）。
5. **无效的深度值**：
   * 命令：`crawl-url <server_url> --depth 0`
   * 预期：退出码为 `1`，`stderr` 包含 `depth` 校验提示（必须大于 0）。

#### HTTP 响应异常处理 ([crawl_http_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_http_error_test.go))
* **场景设计**：
  * `/` -> 包含 `/err404` (返回 404 响应) 和 `/err500` (返回 500 响应)。
* **校验逻辑**：
  * 命令行正常运行完成，退出码为 `0`（局部页面加载失败不中断整体爬取，具有鲁棒性）。
  * `stdout` 仅包含 `/`，不能包含 `/err404` 和 `/err500`。
  * `stderr` 打印相应的失败 debug 日志（若启用 `--debug`）。

---

## 3. 实施规划

1. **Step 1**：创建 [tests/main_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/main_test.go)，实现测试前的自动化编译和通用 `runCLI` 辅助方法。
2. **Step 2**：编写正向测试用例：
   * [tests/crawl_basic_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_basic_test.go)
   * [tests/crawl_limit_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_limit_test.go)
   * [tests/crawl_depth_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_depth_test.go)
   * [tests/crawl_exclude_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_exclude_test.go)
   * [tests/crawl_media_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_media_test.go)
   * [tests/crawl_output_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_output_test.go)
3. **Step 3**：编写反向与异常测试用例：
   * [tests/crawl_cli_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_cli_error_test.go)
   * [tests/crawl_http_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_http_error_test.go)
4. **Step 4**：运行 `go test -v ./tests/...` 对所有 E2E 用例进行全面验证。
