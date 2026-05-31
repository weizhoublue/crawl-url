# crawl-url CLI E2E 测试实施计划

本计划定义了 `crawl-url` CLI 工具端到端测试用例的实施步骤、修改范围与验证标准。

## 1. 实施步骤与任务列表

### 任务 1: 创建测试脚手架 ([tests/main_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/main_test.go))
* **修改范围**：创建 `tests` 目录及 [main_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/main_test.go) 文件。
* **实现内容**：
  * 在 `TestMain` 中，在所有测试执行前通过 `go build` 编译 `cmd/crawl-url` 为临时二进制文件。
  * 提供 `runCLI` 辅助方法捕获 stdout、stderr 和退出码。
* **验证方式**：在项目根目录执行 `go test -c ./tests` 检查是否能编译成功。

### 任务 2: 实现基础递归与前缀过滤测试 ([tests/crawl_basic_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_basic_test.go))
* **修改范围**：新建 [crawl_basic_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_basic_test.go)。
* **实现内容**：
  * 使用 `httptest.NewServer` 模拟基本的三级页面结构以及外部域名链接。
  * 调用 `runCLI` 并不带额外限制参数，验证默认的递归爬取行为。
  * 断言输出中精准包含站内页面，且不包含外站页面。
* **验证方式**：`go test -v ./tests -run TestCrawlBasic`

### 任务 3: 实现 URL 限制与层级限制测试 ([tests/crawl_limit_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_limit_test.go) 与 [tests/crawl_depth_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_depth_test.go))
* **修改范围**：新建 [crawl_limit_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_limit_test.go) 和 [crawl_depth_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_depth_test.go)。
* **实现内容**：
  * `crawl_limit_test.go`：设定包含多个链接的网页，使用 `--url-limit` 限制爬取数量，断言输出行数等于限制的 URL 数量。
  * `crawl_depth_test.go`：设定具有明确层级深度（D0, D1, D2, D3）的测试网页。使用 `--depth` 限制爬取深度，断言没有超出预期的深度页面被访问或输出。
* **验证方式**：`go test -v ./tests -run "TestCrawlLimit|TestCrawlDepth"`

### 任务 4: 实现排除前缀与媒体资源提取测试 ([tests/crawl_exclude_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_exclude_test.go) 与 [tests/crawl_media_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_media_test.go))
* **修改范围**：新建 [crawl_exclude_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_exclude_test.go) 和 [crawl_media_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_media_test.go)。
* **实现内容**：
  * `crawl_exclude_test.go`：验证 `--exclude-prefix` 能够阻断特定前缀的页面访问与输出。
  * `crawl_media_test.go`：设计包含 `<img>` 和 `<video>` 的网页，对比默认状态与开启 `--image`、`--video` / `--vedio` 参数后的输出结果。
* **验证方式**：`go test -v ./tests -run "TestCrawlExclude|TestCrawlMedia"`

### 任务 5: 实现本地 HTML 保存测试 ([tests/crawl_output_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_output_test.go))
* **修改范围**：新建 [crawl_output_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_output_test.go)。
* **实现内容**：
  * 使用 `--output-dir` 参数指向一个临时测试目录。
  * 执行爬取后，检查该目录下是否生成了与爬取路径相对应的 HTML 文件，内容需与 Mock 服务提供的一致。
* **验证方式**：`go test -v ./tests -run TestCrawlOutput`

### 任务 6: 实现命令行异常与 HTTP 服务器错误测试 ([tests/crawl_cli_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_cli_error_test.go) 与 [tests/crawl_http_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_http_error_test.go))
* **修改范围**：新建 [crawl_cli_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_cli_error_test.go) 和 [crawl_http_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_http_error_test.go)。
* **实现内容**：
  * `crawl_cli_error_test.go`：校验无种子、短参数、未知参数、无效的 `--workers` 与 `--depth` 数值输入。断言退出码应匹配（2 或 1），且 `stderr` 含有指定错误信息。
  * `crawl_http_error_test.go`：Mock Server 返回 404 与 500。断言 CLI 能容错处理，正常退出（退出码为 0），且失败页面不在 `stdout` 中输出。
* **验证方式**：`go test -v ./tests -run "TestCrawlCLIError|TestCrawlHTTPError"`

### 任务 7: 完整运行与提交
* **修改范围**：测试代码整体。
* **实现内容**：
  * 执行所有测试。
  * 将所有测试文件提交到 Git 分支，并使用 `git commit -s -S` 签名提交。
* **验证方式**：`go test -v ./tests/...`

---

## 2. 代码回滚与提交规范

* 每次任务完成后，必须使用 `go test` 进行回归测试，确保新加文件没有导致已有测试发生 Regression。
* 提交说明格式：`test: implement cli end-to-end test cases` 并附带签名。
