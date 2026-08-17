# Bug 是什么

客户端取消集合查询请求后，数据库工作没有收到取消信号，仍等待完整的模拟查询延迟。

# 如何触发

在 bug base 中运行：

```bash
CGO_ENABLED=0 go test ./grader -run '^TestCanceledCollectionRequestsStopDatabaseWork$' -count=1
```

# 错误信息

四类集合请求均报告取消传播耗时约 250ms，超过 150ms 上限，例如：

```text
request cancellation took 252ms to reach database work
```
