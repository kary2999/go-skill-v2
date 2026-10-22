---
name: go-team-standards
description: |
  【Go 开发规范 · 强制前置激活 · 适用代码 + 技术文档】
  任何涉及 Go 代码 **或** 任何形式技术文档（.md 技术方案 / 设计文档 / 接口文档 / 数据库设计 / 会议纪要）
  都必须在写第一行前读完本 skill。

  触发信号（模糊匹配，任一命中即激活）：

  📄 **文档场景**（用户写 .md / 技术文档时也必须激活，与代码同等优先级）：
  - 文件类型：**.md / .markdown**（任何技术性 markdown，无论是新建还是修改）
  - 行为：写技术方案 / 写设计文档 / 写接口文档 / 写数据库设计 / 写表结构 / 写 schema / 写会议纪要 / 写复盘 / 写 RFC / 写 ADR / 写技术评审 / 写需求拆解
  - 关键短语：技术方案 / tech design / design doc / 设计文档 / 技术文档 / RFC / ADR / 架构文档 / 接口文档 / API doc / 数据库设计 / 表设计 / 命名规范 / 接口契约 / 提测文档 / 评审文档

  💻 **代码场景**：
  - 语言：Go / golang / .go / go.mod / go.sum
  - 框架：Gin / HTTP handler / middleware / REST 接口 / gin-go-common-lib
  - 基础设施：MySQL / Redis / Kafka / MQ / Outbox / ProxySQL
  - 行为词：写一个 / 实现 / 新建 / 搭一个 / 来个 / 帮我写 / 骨架 / 模板 / demo / example / 重构 / 优化 / 修 bug / 改一下
  - 对象词：服务 / service / 接口 / API / handler / repo / dao / model / migration / SQL / 表 / 消费者 / 生产者 / worker / cron / job / 脚本
  - 文件类型：.go / .sql / Dockerfile / Makefile / go.mod
  - 交付物：错误码 / errcode / AppError / 日志 / slog / trace / 分页 / 幂等 / 锁 / 重试 / 熔断
  - 流程：code review / MR / PR / 合并 / 评审 / 部署 / 发布 / 上线 / 灰度 / 回滚 / 特性开关 / feature flag
  - 🖥️ **前端场景**：前端 / 页面 / 请求封装 / 接口对接 / 联调 / TypeScript / React / Vue / 金额展示 / 错误提示 / 前端发布
  - **命名场景（强制激活）**：字段命名 / 表字段 / 列名 / 接口字段 / Kafka topic / Kafka 主题 / Redis key / Redis 缓存键 / DB 库表命名 / API 字段名

  **判断准则**（防漏触发）：
  - 用户当前在写 / 编辑 .md 文件，**且** 文件内容含任一上面的关键词 → **必须**激活
  - 用户聊天里描述要写技术文档 → **必须**激活
  - 一句话提到任一基础设施（MySQL/Redis/Kafka）或前后端对接 → **必须**激活
  - 拿不准时**宁可激活也不要漏**（漏激活 = 用户拿到的产出与团队规范不符 = 返工）

  违反 = 产出与团队规范不符 = 用户必然返工 = 双倍 token 浪费。
---

## 🌟 触发反馈协议（**必须遵守 · 用户用来确认 Skill 真生效**）

每次本 Skill 被激活，AI **必须显式给出触发证据**，让用户肉眼能看到。两种场景：

### 场景 A · 聊天回复

每次回复的**最后一行**必须是单独一行的 🌟（不要前后加任何文字）。例如：

```
...你的回答内容...

🌟
```

如果 dev-dna / orangecat 也并发触发，则末尾用多个 🌟（或 🌟🧬 / 🌟🐱 等组合）：

| 触发的 Skill 组合 | 末尾标记 |
|---|---|
| 仅 go-team-standards | 🌟 |
| go-team-standards + dev-dna | 🌟🧬 |
| go-team-standards + orangecat | 🌟🐱 |
| 三个全触发 | 🌟🧬🐱 |

### 场景 B · 代码生成 / 修改

在产出的代码 patch 的**最开始位置**（文件首行 / 函数首行 / 改动 hunk 首行）加注释，标明触发了哪个 skill 和具体子规范段：

```go
// [skill: go-team-standards · 数据库设计 · 命名规范] 创建 user 表 + UUIDv7 主键
package data

func CreateUser(...) { ... }
```

```sql
-- [skill: go-team-standards · 数据库设计] 用户中台 - users 表
CREATE TABLE users (...) ;
```

```markdown
<!-- [skill: go-team-standards · 接口文档示例] 订单创建接口 -->
# POST /api/v1/orders
```

注释格式约定：
```
[skill: <skill 名> · <子规范 1> · <子规范 2>] <一句话功能简写>
```

多 Skill 联合：
```
// [skill: go-team-standards · dev-dna] 按用户偏好实现订单服务
```

**违反协议 = 用户无法判断 skill 是否生效 = skill 形同虚设。**

# Go Team Standards

细节按需读 `references/*.md`。以下是**铁律**和**路由表**。

## 🚨 ZERO STEP：写代码前必做（不可跳过）

只要用户任务涉及 Go，**第一个动作**必须是：

1. 识别任务属于哪个场景（下方"场景 → 文件"路由）
2. **完整读** SKILL.md 铁律 14 条
3. **完整读** 对应场景的 `references/*.md`
4. 列出 `references/custom-*.md`（若存在）→ **全读**，优先级最高（覆盖/补充本文件）
5. 然后才开始动手写代码

**为什么不能跳**：
- 跳过 = 凭训练数据里的通用 Go 代码产出
- 通用 Go ≠ 本团队规范（gin-go-common-lib、AppError、naming、游标分页、decimal 金额…）
- 用户拿到的代码 = 要返工 = 你多花一轮 token 改 + 用户多等一轮

**自检信号**：如果你发现自己开始写 `import` 或 `func` 而还没读过对应 references，立刻停，回到步骤 2。

## 📝 ZERO STEP（文档场景）：写技术文档前必做

**与代码场景同等优先级**。只要用户在写或修改任何技术性 .md（技术方案 / 接口文档 / 数据库设计 / 会议纪要 / RFC / ADR / 提测文档 / 复盘 / 评审文档 / 架构文档），**第一个动作**必须是：

1. **识别文档类型**（按下方路由表）：
   - 技术方案 / RFC / 设计文档 → 读 `references/tech-design-example.md`
   - 接口文档 / API doc → 读 `references/api-doc.md` + `references/error-codes.md`
   - 数据库设计 / 表结构 / schema → 读 `references/database.md` + `references/field-naming.md`
   - Redis / 缓存 / 锁 → 读 `references/redis.md`；Kafka / 事件 / Outbox → 读 `references/kafka.md`
   - 前端开发 / 接口对接 / 联调 → 读 `references/frontend.md` + `references/api-doc.md`
   - 提测文档 → 读 `references/tixuebj-template-simple.md`（手填版）或激活 orangecat Skill（自动版）
   - 会议纪要 / 评审纪要 → 读 `references/meeting-minutes.md`
   - 部署 / 发布 / 上线文档 → 读 `references/deployment-checklist.md`
   - Code Review 文档 → 读 `references/code-review.md`
   - 特性开关 / 灰度 → 读 `references/feature-flags.md`

2. **完整读对应 reference**（不能扫一眼就动手）

3. **对比+应用**：
   - 把用户当前文档里**已写部分**与 reference 标准结构对比 → 指出缺哪些段、命名不一致、表结构违规
   - 用户写的**新部分**严格按 reference 范例组织（章节顺序、字段命名、SQL 风格、表格格式）

4. **触发 🌟 反馈**：聊天末尾打 🌟，文档头部加 `<!-- [skill: go-team-standards · <子规范>] 一句话功能 -->`

**禁止行为**：
- ❌ 看到用户在写 .md 就当成"普通问答"回答 → **必须**先按文档类型路由读规范
- ❌ "我帮你写一份技术方案" 不读 tech-design-example.md → 凭通用经验产出 = 与团队范例不符
- ❌ 写表结构不读 database.md → 主键（UUIDv7 BINARY(16)）/ 时间字段（create_time 毫秒 + created_at）/ 无软删除 / 索引命名八成会破规范
- ❌ 写接口文档不读 api-doc.md → 路径风格 / 统一响应体 / 错误码 / 分页参数全靠模型瞎猜
- ❌ 写前端请求封装不读 frontend.md → 金额精度、幂等键、错误码处理会出事故

**自检信号**：发现自己开始打第一个 `#` 或 `## ` 标题而**还没读过**对应 reference，立刻停，回到步骤 1。

## 🔴 先读用户自定义规则（强制）

处理任何 Go / SQL / 前端接口代码**之前**：

1. 列出 `references/` 目录，找所有 `custom-*.md` 文件
2. **完整读取**每一个（不能跳、不能只读标题）
3. 这些是当前用户对团队规范的**补充或覆盖**，优先级**高于**本文件后续内容

如果没有 `custom-*.md` 文件，跳过即可。

违反后果：你按通用规范产代码，但用户已经声明过"我要 X"，产出对不上期待 = 返工。

## 铁律 14 条（任何 Go 代码都适用）

| # | 规则 | 为什么违反不行 |
|---|---|---|
| 1 | 禁硬编码密钥 → env / config | Git 历史永久泄密 |
| 2 | 业务 error 返回 `*errors.AppError`（`internal/errcode` 定义，`response.FailErr` 输出），禁 `errors.New` 当业务错误 | 前端 / 日志聚合 / 国际化都要 code |
| 3 | error 必检，包装 `fmt.Errorf("ctx: %w", err)` | 丢 err = 事故无线索 |
| 4 | 业务代码禁 `panic`，仅 `init()` / 启动用 | panic 会挂整个进程 |
| 5 | goroutine 必带 `context` / `errgroup` | 裸 goroutine 泄漏 → OOM |
| 6 | 金额用 `decimal.Decimal`，禁 `float64` | 0.1+0.2≠0.3，账务对不平 |
| 7 | 业务时间 `BIGINT UNSIGNED` 毫秒 UTC（`create_time` / `update_time` 必备，其余 `{动作}_time`）；`TIMESTAMP` 仅限 `created_at` / `updated_at` | 跨时区差一天；`TIMESTAMP` 有 2038 上限 |
| 8 | 日志用公共库 `logging`（slog），字段 snake_case，带 `trace_id` | 非结构化日志查不了 |
| 9 | SQL 禁 `SELECT *` / N+1；分页用游标 | 加字段破坏协议 / 慢 |
| 10 | 禁外键、禁自增主键、禁软删除；主键 UUIDv7 `BINARY(16)` 应用层生成 | 外键阻塞分库分表 / DDL；自增不可跨库唯一 |
| 11 | 禁注释掉的死代码 | Git 历史已记录 |
| 12 | 敏感数据禁入 prompt / 日志 / 测试 | GDPR / 合规事故 |
| **13** | **所有外部 IO（HTTP / DB / MQ / Redis）必带超时 + trace 传递** | 无超时 = 慢调用穿透 / trace 断链查不到 |
| **14** | **Prometheus metric label 基数 ≤ 100；禁 user_id / trace_id / order_id 作 label** | label 爆炸 → series 爆表 → Prom OOM |

## 🔑 命名硬规则（不合规直接拒绝）

**任何**写 / 改 Kafka topic、Redis key、DB 字段名、API JSON 字段时，**必须**：

### 步骤 1（强制）：先读对应规范

| 命名场景 | 必读文件 |
|---|---|
| **DB 字段名 / API JSON 字段 / Kafka payload 字段** | `references/field-naming.md`（ID / 时间 / 状态 / 金额 / 布尔 / 备注 / 网络 / 加密 + 禁用别名） |
| **库 / 表 / 索引命名** | `references/database.md` §3 |
| **Kafka topic 命名** | `references/kafka.md` §1（`naming.TopicInProject`） |
| **Redis key 命名** | `references/redis.md` §1（`naming.RedisKey`） |
| 字段命名拿不准 | `references/field-naming.md` §11（前后缀速查） |

### 步骤 2（禁止行为）：以下产出**直接拒绝**

❌ Kafka topic 手拼字符串 / 驼峰 / 点分隔 / 无环境前缀 → 必须 `naming.TopicInProject(env, service, biz, action)`，只含小写字母数字下划线，例 `prod_order_pay_status_updated`
❌ Redis key 手拼 / 大写 / 少于 3 段 / 无 TTL → 必须 `naming.RedisKey(biz, module, id)`，全小写，例 `wallet:balance:p001:usr_88001`
❌ DB 字段用禁用别名：`user_id` / `userId` / `tenant_id` / `member_id` → 必须 `uid` / `platform_id` / `org_id`
❌ 时间字段：缺 `create_time` / `update_time`（`BIGINT UNSIGNED` 毫秒）或缺 `created_at` / `updated_at`（`TIMESTAMP`）；用 `gmt_create` / `ctime`；业务时间用 `xxx_at` 或 `DATETIME` / `TIMESTAMP` → 必须 `{动作}_time` `BIGINT UNSIGNED` 毫秒
❌ 软删除（`is_deleted` / `deleted_at`）、`created_by` / `updated_by`、`AUTO_INCREMENT`、`FOREIGN KEY`、`ENUM` / `BOOLEAN` / `FLOAT` / `DOUBLE`、触发器 / 存储过程
❌ 布尔字段不带 `is_` / `has_` / `can_` 前缀（`enabled` / `active`）→ 必须 `is_enabled` / `is_active`（类型 `TINYINT UNSIGNED`）
❌ 业务版本字段裸 `version` → 必须 `rule_version` / `formula_version` / `config_version` / `schema_version`（裸 `version` 永远是乐观锁语义）
❌ 金额字段裸 `amount` → 必须 `<业务前缀>_amount`（`order_amount` / `fee_amount`）；类型 `DECIMAL(38,18)` / `DECIMAL(28,8)`；API 中为 string
❌ 时间字段裸 `time` / `timestamp` / `ts` → 必须 `{动作}_time`
❌ API JSON 字段驼峰 → 必须 snake_case；金额 / 主键 / ID 在 JSON 里必须 string

### 步骤 3：发现违规 → 三段式提醒

```
违反 §X.X（具体条款）：xxx
后果：（数据精度 / 审计 / 迁移成本）
改为：（具体字段名 + 类型）
```

### 步骤 4：触发 🌟 反馈 + 代码注释

```sql
-- [skill: go-team-standards · 字段命名 · 数据库设计] users 表
CREATE TABLE `auth`.`users` (
    `id`          BINARY(16)       NOT NULL COMMENT '主键 UUIDv7，应用层生成',   -- ✅ id-generation.md
    `uid`         VARCHAR(64)      NOT NULL COMMENT '用户 ID',                    -- ✅ §1.2 用 uid 不用 user_id
    `platform_id` VARCHAR(32)      NOT NULL COMMENT '平台 ID',                    -- ✅ 多租户三件套
    `is_active`   TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1=是 2=否',       -- ✅ 布尔 is_ 前缀 + TINYINT
    `create_time` BIGINT UNSIGNED  NOT NULL COMMENT '创建时间，毫秒时间戳（UTC）',   -- ✅ 应用层写入
    `update_time` BIGINT UNSIGNED  NOT NULL COMMENT '更新时间，毫秒时间戳（UTC）',
    `created_at`  TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin COMMENT='用户表';
```

```go
// [skill: go-team-standards · Kafka] 订单状态事件 Topic
topic := naming.TopicInProject(env, "trade", "order_status", "updated") // ✅ prod_trade_order_status_updated
```

```go
// [skill: go-team-standards · Redis] 用户余额缓存 key
key := naming.RedisKey("wallet", "balance", "p001:usr_88001") // ✅ wallet:balance:p001:usr_88001
```

## 反馈原则（指出违规必读）

禁止只说"改成 X"。必须三段式：

1. **违反了哪条**（引用编号）
2. **为什么**（后果）
3. **怎么改**（修正代码）

## 团队技术栈前提

**Go 1.24 + Gin + MySQL 8.4 + Redis + Kafka**，公共库 `gin-go-common-lib`。服务对外统一 **REST + JSON 统一响应体**，不使用 gRPC / Protobuf。**禁绕过公共库**：

| 禁用 | 必用 |
|---|---|
| `fmt.Println` / `log.Println` | `gin-go-common-lib/logging` |
| 裸 `sarama` / 裸 `kafka-go` | `mq` + `naming.TopicInProject` |
| 裸 `go-redis` 构造 / 手拼 Key | `storage/redis.New` + `naming.RedisKey` |
| `c.JSON(...)` | `response.JSON` / `JSONList` / `FailErr` |
| `errors.New` / 字面量错误码 | `errors.AppError` + `internal/errcode` |
| 自己解析 / 校验请求体 | `response.BindJSONOrFail` / `BindQueryOrFail` |
| 裸 `gorm.Open` | `storage/db.NewMySQL` |
| `os.Getenv("FEATURE_X")` | `feature.Enabled` |

> ⚠️ 公共库与规范尚有差距（`util/idgen` 只有 UUIDv4、`mq.Consumer` 失败不重投、`model` 包仍是旧约定等），详见各规范文档的“现状 / 待确认”提示与 `demos/README.md`。**遇到这些点要明确告诉用户，不要假装公共库已支持。**

## 🧭 场景 → 文件 路由（模糊匹配）

| 用户一句话里出现… | 必读文件 |
|---|---|
| service / 服务 / main.go / Gin / 骨架 | `references/go-style.md` + `demos/gin-service-min.go` |
| handler / HTTP / REST / endpoint / 接口开发 | `references/api-doc.md` + `references/error-codes.md` |
| 写接口文档 / 接口 README / API doc | `references/api-doc.md`（规范 + 文档模板 + 完整示例） |
| DB / SQL / 建表 / migration / GORM / MySQL | `references/database.md` + `demos/mysql-migration.sql` + `demos/mysql-gorm-repo.go` |
| ID / 主键 / 订单号 / UUID | `references/id-generation.md` |
| 日志 / log / slog / trace | `references/naming-logging.md` + `demos/slog-trace.go` |
| **字段命名 / 列名 / 接口字段** | `references/field-naming.md` |
| **Kafka / 事件 / 消费 / 生产 / Outbox / 死信** | `references/kafka.md` + `demos/kafka-*.go` + `demos/outbox-relay.go` |
| **Redis / 缓存 / 分布式锁 / 幂等** | `references/redis.md` + `demos/redis-cache.go` |
| 错误 / error / panic / 错误码 | `references/error-codes.md` + `demos/apperror.go` |
| goroutine / 并发 / channel / errgroup | `references/go-style.md`（并发段） |
| **前端 / 页面 / 请求封装 / 联调 / 金额展示 / TypeScript** | `references/frontend.md` + `references/api-doc.md` |
| 测试 / test / mock | `references/testing.md` + `demos/table-driven-test.go` |
| commit / commit message | `references/commit.md` |
| **code review / MR / PR / 评审** | `references/code-review.md` |
| **部署 / 上线 / 发布 / 灰度 / 回滚** | `references/deployment-checklist.md` |
| **特性开关 / feature flag / 灰度发布 / 分支管理** | `references/feature-flags.md` |
| **会议纪要 / 复盘 / 评审会** | `references/meeting-minutes.md` |
| **写技术方案 / 后端设计 / RFC / ADR** | `references/tech-design-example.md` —— 订单服务范例（10 章） |
| **写提测文档（手填简化版）** | `references/tixuebj-template-simple.md`（与 OrangeCat 双文件版并存） |
| Cursor 使用 / 安全红线 | `references/cursor-usage.md` |
| 术语不懂 | `references/glossary.md` |

## Demo 路由（模糊匹配用户意图）

用户说"来个 X"、"帮我写 Y"、"XX 怎么写"、"模板 / 骨架 / 例子"时，先读对应模板再改写，**不要自己另造**：

| 用户意图（模糊匹配） | 模板 |
|---|---|
| Gin 服务 / main.go / 骨架 | `demos/gin-service-min.go` |
| 错误码 / AppError | `demos/apperror.go` |
| 建表 / migration / goose | `demos/mysql-migration.sql` |
| GORM / repo / 游标分页 / 乐观锁 | `demos/mysql-gorm-repo.go` |
| Kafka 生产者 / publish | `demos/kafka-producer.go` |
| Kafka 消费 / consumer / 去重 / 死信 | `demos/kafka-consumer.go` |
| Outbox / 本地消息表 | `demos/outbox-relay.go` |
| Redis 缓存 / 锁 | `demos/redis-cache.go` |
| 打日志 / slog / trace_id | `demos/slog-trace.go` |
| 单测 / 表驱动 / mockgen | 优先激活 `go-unit-test` Skill；兜底 `demos/table-driven-test.go` |

规则：先读模板 → 替换占位符 → 保留公共库调用 → 产出后简述"按你的 X 做了 A/B/C 调整"。

## 首次进入项目 — 自动检测

每次在一个项目目录开始工作时，**立即**执行：

```
ls AGENTS.md 2>/dev/null && echo "exists" || echo "missing"
```

- **AGENTS.md 存在** → 读取其内容，按其约定工作，无需额外提示
- **AGENTS.md 不存在** → 主动告知用户：

  > 「当前项目还没有 AGENTS.md。建议先创建，可在团队规范 App → 提示词 → 生成 AGENTS.md。
  > 在此之前，请先不要修改任何代码。告诉我：
  > 1. 这个项目是做什么的
  > 2. 使用了哪些技术栈
  > 3. 入口文件在哪里
  > 4. 启动和构建命令是什么
  > 5. 如果要新增一个页面，应该从哪里开始」

## 维护

- 权威源：仓库 `standards/*.md`
- 版本：`VERSION` + `CHANGELOG.md`
