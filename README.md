# 家庭共享购物清单 API

基于 Go 1.22、MySQL 8、JWT 和 bcrypt 实现的家庭共享购物清单后端。一个清单可有多名成员，成员通过邀请码加入；成员可以添加、修改、勾选和删除商品，并查看每条商品的最后修改人和修改时间。

## 技术栈

- Go 1.22 + `net/http`
- MySQL 8
- `database/sql` + `github.com/go-sql-driver/mysql`
- `github.com/golang-jwt/jwt/v5`
- `golang.org/x/crypto/bcrypt`
- Docker Compose

## 目录结构

```text
.
├── api/                    # HTTP 路由组装
├── cmd/server/             # 服务入口
├── configs/                # 配置示例
├── internal/
│   ├── config/             # 环境变量配置
│   ├── database/           # 连接、等待、迁移、种子数据
│   ├── middleware/         # JWT 认证、清单成员/所有者授权
│   ├── invite/             # 邀请码 handler/service/repository/model
│   ├── item/               # 商品 handler/service/repository/model
│   ├── list/               # 清单 handler/service/repository/model
│   ├── member/             # 成员 handler/service/repository/model
│   └── user/               # 用户 handler/service/repository/model
├── migrations/             # 自动执行的 SQL 迁移
├── pkg/response/           # 统一 JSON 响应
├── Dockerfile
└── docker-compose.yml
```

## 快速启动

```bash
cp .env.example .env
docker compose up -d --build
```

服务映射到宿主端口 `18091`：

```bash
curl http://127.0.0.1:18091/healthz
```

应用启动时会等待 MySQL 就绪、自动执行 `migrations/*.sql`，并创建以下演示数据：

| 用户名 | 密码 | 说明 |
| --- | --- | --- |
| `demo` | `demo12345` | 演示清单所有者 |
| `alice` | `alice12345` | 演示清单成员 |
| `bob` | `bob12345` | 演示清单成员 |

演示清单 `家庭日常采购` 中已有 `牛奶` 和 `鸡蛋`。

## API 摘要

所有业务接口除 `/healthz` 和 `/api/v1/auth/*` 外，都需要请求头：

```http
Authorization: Bearer <JWT>
```

| 方法 | 路径 | 权限 |
| --- | --- | --- |
| POST | `/api/v1/auth/register` | 公开 |
| POST | `/api/v1/auth/login` | 公开 |
| GET | `/api/v1/me` | 登录用户 |
| POST | `/api/v1/lists` | 登录用户 |
| GET | `/api/v1/lists` | 登录用户 |
| GET | `/api/v1/lists/{list_id}` | 清单成员 |
| PATCH | `/api/v1/lists/{list_id}` | 清单所有者 |
| DELETE | `/api/v1/lists/{list_id}` | 清单所有者 |
| GET | `/api/v1/lists/{list_id}/members` | 清单成员 |
| DELETE | `/api/v1/lists/{list_id}/members/{member_id}` | 清单所有者 |
| GET | `/api/v1/lists/{list_id}/invites` | 清单成员 |
| POST | `/api/v1/lists/{list_id}/invites` | 清单所有者 |
| POST | `/api/v1/invites/join` | 登录用户 |
| GET | `/api/v1/lists/{list_id}/items` | 清单成员 |
| POST | `/api/v1/lists/{list_id}/items` | 清单成员 |
| GET | `/api/v1/lists/{list_id}/items/{item_id}` | 清单成员 |
| PATCH | `/api/v1/lists/{list_id}/items/{item_id}` | 清单成员 |
| DELETE | `/api/v1/lists/{list_id}/items/{item_id}` | 清单成员 |

加入清单必须先登录并获得有效邀请码；没有邀请码且不是清单成员的用户访问清单接口会得到 `403 Forbidden`。

## 配置

环境变量见 `.env.example`：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `SERVER_PORT` | `8080` | 容器内服务端口 |
| `DB_HOST` | `mysql` | MySQL 主机 |
| `DB_PORT` | `3306` | MySQL 端口 |
| `DB_USER` | `shopping` | MySQL 用户 |
| `DB_PASSWORD` | `shopping_secret` | MySQL 密码 |
| `DB_NAME` | `family_shopping` | 数据库名 |
| `JWT_SECRET` | 开发默认值 | JWT 签名密钥，生产必须替换 |
| `TOKEN_TTL_HOURS` | `24` | Token 有效期 |
| `MIGRATIONS_DIR` | `./migrations` | 迁移目录 |

## 清理

```bash
docker compose down -v
```
