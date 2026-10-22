# Demos — 轻量最佳实践片段（Go + Gin + MySQL 8.4 + Redis + Kafka）

这不是完整可跑项目，是**能直接复制到服务脚手架里的代码片段**。每个 demo 聚焦一个场景，优先调用公共库 `gin-go-common-lib` 的封装，**不绕开团队基础设施**。

> **每个文件是独立片段，不能把整个目录一起编译**（各文件的 `package` 不同，对应它们在真实项目里的位置）。

## 何时用
用户说类似“帮我写个 kafka 生产者”、“来个 redis 缓存 / 锁”、“建表模板”、“新建 gin 服务”、“错误码怎么用”……AI 应该识别意图并参考对应 demo 产出代码。**模糊匹配即可**，不要求关键词完全一致。

## Demo 索引

| 文件 | 真实位置 | 场景 | 触发关键词（模糊） |
|---|---|---|---|
| `gin-service-min.go` | `cmd/trade/main.go` + handler / biz | 最小 Gin 服务骨架：中间件、统一响应、幂等、MySQL / Redis 初始化、优雅退出 | gin / 新服务 / 骨架 / 启动入口 / handler |
| `apperror.go` | `internal/errcode/` | 业务错误码定义（`AppError`）与用法 | 错误码 / errcode / AppError / return error |
| `mysql-migration.sql` | `migrations/` | goose 迁移模板：必备字段、UUIDv7 主键、索引、COMMENT | 建表 / 迁移 / migration / goose / schema |
| `mysql-gorm-repo.go` | `internal/data/order_repo.go` | GORM 仓储：ctx 超时、`BINARY(16)`、乐观锁、游标分页、物理删除、1062 处理 | gorm / 仓储 / repo / crud / 查询 / 分页 |
| `kafka-producer.go` | `internal/data/` | 发布事件：`naming` 构造 Topic、统一信封、带 header | kafka / 生产者 / 发消息 / producer / topic |
| `outbox-relay.go` | `internal/job/` | Outbox relay：`SKIP LOCKED` 取数、保序发送、失败重试 | outbox / 本地消息表 / relay / 最终一致 |
| `kafka-consumer.go` | `internal/job/` | 消费者：去重表同事务、handler 内重试 + 死信 | kafka / 消费者 / consumer / 幂等消费 / 死信 |
| `redis-cache.go` | `internal/data/` | Cache-Aside、singleflight、TTL 抖动、分布式锁（Lua 释放） | redis / 缓存 / 锁 / lock / 击穿 |
| `slog-trace.go` | — | `logging` 日志 + `trace_id` 用法 | 日志 / slog / trace_id / 打日志 |
| `table-driven-test.go` | `*_test.go` | 表驱动单测 + mockgen 骨架 | 单测 / 测试 / 表驱动 / mock / 覆盖率 |

## 验证情况（诚实说明）

* 以下 demo 已在**临时模块里对公共库 `gin-go-common-lib` 编译通过（`go vet`）**：`gin-service-min.go`、`apperror.go`、`mysql-gorm-repo.go`、`kafka-producer.go`、`kafka-consumer.go`、`outbox-relay.go`、`redis-cache.go`、`slog-trace.go`
* 以下已用**真实 MySQL 8.4.11 / Redis 7.2** 跑过集成测试：`mysql-migration.sql`（建表）、`mysql-gorm-repo.go`（建单 / 冲突 / 读回精度 / 乐观锁 / `clientFoundRows` / 游标分页 / 删除 / CHECK）、`outbox-relay.go`（保序与失败重试）、`kafka-consumer.go` 的 `process`（重复投递只处理一次）、`redis-cache.go`（singleflight / TTL 抖动 / 锁互斥与 token 释放）
* **未验证**：`gin-service-min.go` 没有实际启动并请求；`kafka-producer.go` 与 `kafka-consumer.go` 的 `Run` / `Publish` **没有连真实 Kafka**；`table-driven-test.go` 是模板，未编译

## 使用约定

1. import 路径按公共库 `gin-go-common-lib` 填写；若你团队的 module path 不同，复制后整体替换。
2. 所有 demo 都假设 `context.Context` 已带 `trace_id`（上游 `middleware.TraceID` 已注入）。
3. 命名参数（订单、交易对等）请按业务替换；类型和方法签名是模板的重点。
4. **demo 里出现的公共库缺口**（需先补公共库再按规范落地）：
   * `BinaryUUID`（`BINARY(16)` 读写）在 demo 里自带，应沉淀进公共库
   * `util/idgen` 目前只有 UUIDv4，demo 里直接用 `uuid.NewV7`，真实项目应走 `idgen`（见 `id-generation.md`）
   * `mq.Consumer` 处理失败不会重投（见 `kafka.md` §4.2），demo 已在 handler 内自行重试 + 死信
   * 公共库的 `model` 包仍是旧约定（`created_by` / `deleted_at` / `CreatedAt time.Time`），**与新规范不一致，不要使用**；demo 自带符合新规范的模型字段
5. 不要把 `[]byte` 直接当 GORM 的 `?` 参数（会展开成多个参数）：用 `driver.Valuer` 类型，或 `UUID_TO_BIN(?)` 配字符串。
