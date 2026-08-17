# Bug 是什么

数据库查询错误经过 repository 返回后丢失 unwrap 链，上层只能看到错误文本。

# 如何触发

在 bug base 中运行：

```bash
CGO_ENABLED=0 go test ./grader -run '^TestDatabaseErrorsRemainInspectable$' -count=1
```

# 错误信息

```text
database error chain lost: 查询商品失败: context deadline exceeded
```
