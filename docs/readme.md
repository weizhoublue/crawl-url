# crawl-url

文档站 URL 递归抓取工具（Go 实现）。由 `../crawUrl/craw_url.py` 重构而来，CLI 与输出语义兼容。

## 构建

```bash
cd my_dev_tools/crawl-url
go build -o crawl-url ./cmd/crawl-url
```

## 用法

```bash
# 递归抓取整个文档站前缀
./crawl-url https://docs.cilium.io/en/stable/ 2>/dev/null

# 最多 100 个 URL（选项使用 -- 长选项格式）
./crawl-url https://docs.cilium.io/en/stable/ --url-limit 100 2>/dev/null

# 8 个并发 worker
./crawl-url https://docs.cilium.io/en/stable/ --workers 8 --url-limit 50 2>/dev/null

# 限制深度：种子=0，depth 2 = 种子 + 一层子链接
./crawl-url https://clerk.com/docs --depth 2 --url-limit 50

# 帮助
./crawl-url --help

# 排除前缀、保存 HTML、调试
./crawl-url https://example.com/docs/ \
  --exclude-prefix https://example.com/docs/_static/ \
  --output-dir ./saved-pages \
  --debug
```



