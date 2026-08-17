# Bug 是什么

请求解码只消费第一个 JSON 值，没有确认请求体已经到达 EOF，因此尾随 JSON 文档被静默忽略。

# 如何触发

在 bug base 中运行：

```bash
CGO_ENABLED=0 go test ./internal/item/handler ./internal/list/handler ./internal/invite/handler ./internal/member/handler -run '^TestRequestBodyRejectsTrailingDocument$' -count=1
```

# 错误信息

```text
expected a request containing two JSON documents to be rejected
```
