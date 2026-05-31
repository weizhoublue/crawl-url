# crawl-url CLI 端到端 (E2E) 测试用例说明

本目录存放 `crawl-url` CLI 工具的端到端集成测试代码，主要针对各种命令行参数及边界表现进行自动化测试。

## 1. 测试框架设计

* **[main_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/main_test.go)**
  * **TestMain**：在运行测试前，自动将宿主 CLI（`cmd/crawl-url`）编译成临时二进制文件，避免每个测试用例重复编译，保证极高的执行效率。测试结束时，自动清理临时构建产物。
  * **runCLI**：通用的子进程启动与输出拦截工具，可捕获 stdout、stderr 及进程退出状态码。

## 2. 用例设计列表

### 2.1 正向测试用例

* **[crawl_basic_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_basic_test.go)**
  * **TestCrawlBasic**：模拟多级页面结构。验证 CLI 递归抓取能力，并断言非种子前缀下的外部链接不会被误抓取。
* **[crawl_limit_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_limit_test.go)**
  * **TestCrawlLimit**：使用 `--url-limit 2` 限制爬取。断言输出的成功爬取 URL 数量（含种子）精确为 2。
* **[crawl_depth_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_depth_test.go)**
  * **TestCrawlDepth**：定义层级明显的测试网页（D0 -> D1 -> D2）。使用 `--depth 2` 运行，断言只抓取 D0（种子）和 D1 层级，D2 层级不被爬取或输出。
* **[crawl_exclude_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_exclude_test.go)**
  * **TestCrawlExclude**：利用 `--exclude-prefix` 排除特定路径下的资源，验证过滤的正确性。
* **[crawl_media_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_media_test.go)**
  * **TestCrawlMediaDefault**：验证默认不抓取图片与视频。
  * **TestCrawlMediaEnabled**：验证配置 `--image --video` 后正确捕获媒体链接。
  * **TestCrawlMediaAlias**：验证拼写别名 `--vedio` 正确映射抓取视频链接。
* **[crawl_output_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_output_test.go)**
  * **TestCrawlOutput**：验证使用 `--output-dir` 参数时，本地能生成并写入正确的 HTML 文件，文件名和路径结构能与内部规则一致。
* **[crawl_scale_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_scale_test.go)**
  * **TestCrawlScaleFull**：构建最大深度为 4 层的二叉树，总计 31 个动态页面。无限制抓取时，断言输出的 URL 数量精确等于 31。
  * **TestCrawlScaleWithLimit**：对 31 个页面的大拓扑加设 `--url-limit 15`，断言输出数量精准为 15。
  * **TestCrawlScaleWithDepth**：对 31 个页面的大拓扑加设 `--depth 3`（抓取深度 0, 1, 2），断言仅输出 7 个节点，且无更深层页面被抓取。

### 2.2 反向与异常用例

* **[crawl_cli_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_cli_error_test.go)**
  * **TestCrawlCLIErrorMissingSeed**：验证缺少种子 URL 时以退出码 `2` 失败，并提示错误。
  * **TestCrawlCLIErrorShortFlag**：验证使用单减号短选项（如 `-workers`）时以退出码 `2` 失败，提示使用长选项。
  * **TestCrawlCLIErrorUnknownFlag**：验证使用未知参数时以退出码 `2` 失败。
  * **TestCrawlCLIErrorInvalidWorkers**：验证输入非法并发数（如 `--workers -1`）时以退出码 `1` 失败。
  * **TestCrawlCLIErrorInvalidDepth**：验证输入非法深度限制（如 `--depth 0`）时以退出码 `1` 失败。
* **[crawl_http_error_test.go](file:///Users/weizhoublue/Documents/git/crawl-url/tests/crawl_http_error_test.go)**
  * **TestCrawlHTTPErrorTolerance**：网页中包含 404 及 500 的子链接。验证 CLI 具备极强的容错性能，能够正常以状态码 `0` 退出，且失败的页面链接绝对不会输出到 stdout 中，错误信息会记录在调试输出（`stderr`）里。

## 3. 运行测试

在根目录下使用以下命令运行此测试套件：

```bash
go test -v ./tests/...
```
