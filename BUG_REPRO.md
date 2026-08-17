# Bug 是什么

集合查询在行扫描提前失败时没有释放结果集，可能长期占用数据库连接。

# 如何触发

在 bug base 中运行：

```bash
CGO_ENABLED=0 go test ./grader -run '^TestCollectionQueriesCloseRowsAfterDecodeFailure$' -count=1
```

# 错误信息

```text
expected query rows to be closed, but it was not
```
