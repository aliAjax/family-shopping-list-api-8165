# Bug 是什么

四类集合查询在结果为空时返回 nil slice，JSON 编码结果为 null 而不是数组。

# 如何触发

在 bug base 中运行：

```bash
CGO_ENABLED=0 go test ./grader -run '^TestEmptyCollectionsUseJSONArray$' -count=1
```

# 错误信息

```text
empty collection encoded as null, want []
```
