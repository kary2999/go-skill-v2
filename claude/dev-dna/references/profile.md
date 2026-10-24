---
title: "我的开发档案"
version: "2.0.0"
last_modified: "2026-10-24"
---

# 🧬 我的开发档案（dev-dna profile）

> 这是默认模板。在 Team Standards App「🧬 我的开发 DNA」卡片里编辑后会覆盖此文件。
> AI 读这份档案后，回答任何技术问题时**必须**遵循下方偏好。

---

## 1. 基本信息

- **角色**：（如：Go 后端工程师 / 微服务架构师 / Tech Lead）___
- **主栈**：（如：Go 1.24 · Gin · GORM · MySQL 8.4 · Redis · Kafka）___
- **次栈**：（次要语言 / 框架，如：TypeScript · React）___
- **团队**：___
- **业务领域**：___（如：交易系统 / 数据中台 / 风控）

## 2. 编码偏好（强 → 弱）

> 描述你**主动选择**的风格。AI 应该按此产出代码。

- **错误处理**：必须包装（`fmt.Errorf("ctx: %w", err)`）+ 业务错误走 `AppError` + `internal/errcode`；禁裸 errors.New
- **命名**：snake_case for SQL columns / camelCase for Go vars / kebab-case for k8s resources
- **测试**：表驱动 + mockgen，禁手写 mock
- **注释**：godoc 风格（`// 函数名 does ...`）+ 中文业务注释
- **代码组织**：interface 先行；构造函数手动注入；handler / biz / data 三层
- **金额**：必须 decimal.Decimal；DB 加密货币 DECIMAL(38,18)、法币 DECIMAL(28,8)；JSON 用字符串
- **日志**：slog + snake_case key + trace_id

（其他自由补充：）
- ___
- ___

## 3. 反偏好（不喜欢看到的代码）

> AI 看到这些应该主动指出 + 改成你偏好的写法。

- ❌ panic 在业务代码
- ❌ SELECT * / OFFSET 大翻页
- ❌ fmt.Println 当日志
- ❌ float64 存金额
- ❌ 注释掉的死代码 / TODO 不带日期 + 责任人
- ❌ 单测里直连真 DB

（其他自由补充：）
- ❌ ___

## 4. 常用模式（"按我的习惯写"对应的具体套路）

> 用户说"按我习惯实现 X"时，AI 直接照这些套路产出，不要再问。

### 4.1 新建一个 Gin 微服务
- cmd/<svc>/main.go 手动装配（构造函数注入），中间件走 `middleware.Default`
- handler / biz / data 三层；handler 只做解析与校验，业务规则在 biz
- 错误码在 internal/errcode/<biz>.go，先注册再使用

### 4.2 新增数据表
- 先写 goose migration 文件 `migrations/{6位序号}_{库}_{描述}.sql`，一个文件只放一条 DDL
- 两个 ID：`id BIGINT UNSIGNED AUTO_INCREMENT`（库内主键，不外露）+ `oid CHAR(36)`（UUIDv7 业务 ID，带 `uk_{表}_oid`），引用写 `{entity}_oid`
- 时间：`create_time` / `update_time`（BIGINT 毫秒，应用层写）+ `created_at` / `updated_at`（`DATETIME(6)`，库维护）；全库禁用 `TIMESTAMP`
- 无软删除（不设 `deleted_at`）、无 `created_by` / `updated_by`；禁外键

### 4.3 新增 HTTP 接口
- 先写接口文档（REST + JSON 统一响应体 `{code,msg,data,trace_id,timestamp}`），契约确认后再写代码
- handler 调 biz，错误统一 `AppError`，前端按 `code` 国际化
- 写操作带 `X-Idempotency-Key`；金额、ID 一律 string

（自由补充更多模式：）
### 4.x ___

## 5. 工具链 / 环境

- **IDE**：（Cursor / Claude Code / GoLand / VS Code）___
- **终端**：（iTerm2 / Warp / kitty）___
- **shell**：（zsh / bash / fish）___
- **关键 alias**：（`gst=git status` / `k=kubectl` 等）___
- **测试命令**：`go test -race -cover ./...`
- **lint**：`golangci-lint run --timeout=5m`

## 6. 决策口味

- **性能 vs 可读性**：默认___（如：可读性优先，热路径再优化）
- **新依赖引入**：默认___（如：保守，先看是否能用 stdlib）
- **上线节奏**：偏好___（如：周二 / 周四，避开周一周五）
- **评审风格**：___（如：先指铁律，再指细节，最后给修正代码）

## 7. 技术债 / 我自己常犯的错

> 让 AI 主动提醒你别再犯。

- ___（如：我经常忘记 ctx.Done() 的 select）
- ___（如：写 SQL 容易漏 LIMIT）

## 8. 学习方向 / 兴趣

> 让 AI 在合适的话题下推荐进阶资料。

- ___（如：分布式一致性 / Raft / OLAP）
- ___

---

## ✏️ 编辑指南

- 每次重大变化（换栈、换团队、风格演进）→ 更新顶部 `last_modified` + 提 `version`
- 内容尽量**具体可执行**（"我喜欢用 AppError + errcode 返回业务错误"），别太抽象（"我喜欢好的错误处理"）
- 跨电脑迁移：**整个 `~/.claude/skills/dev-dna/` 目录 tar 走**就行
