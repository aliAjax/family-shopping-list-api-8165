# Bug 是什么

集合查询脱离传入的请求 context，且在行扫描提前失败时没有释放结果集；请求结束或取消也无法回收连接。

# 如何触发

在 bug base 中运行：

```bash
CGO_ENABLED=0 go test ./grader -run '^TestCollectionQueriesCloseRowsAfterDecodeFailure$' -count=1
```

# 错误信息

```text
expected query rows to be closed, but it was not
```
