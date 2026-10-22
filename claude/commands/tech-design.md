---
description: 写技术方案（按 tech-design-example.md 10 章范例）
argument-hint: <主题，如 "用户中台 v2 改造">
---

# 写技术方案：$ARGUMENTS

请按下列流程为我产出**技术方案**：

## 步骤

1. **必读**：`~/.claude/skills/go-team-standards/references/tech-design-example.md`（订单服务完整范例，10 章）
2. **必读**：`~/.claude/skills/go-team-standards/references/database.md`（如果方案涉及数据库）
3. **必读**：`~/.claude/skills/go-team-standards/references/api-doc.md`（如果方案涉及接口）
   - 涉及缓存 / 消息：另读 `redis.md` / `kafka.md`；涉及前端：另读 `frontend.md`
4. 按范例的 **10 章结构**输出：
   - 背景与概述 / 目标与原则 / 规范定义
   - 核心业务流程（含失败与补偿）
   - 应用架构 / 系统架构
   - 功能模块设计
   - 接口清单（按 api-doc.md 模板）
   - 数据库设计（按 database.md，含容量估算）
   - 边界与异常处理（含监控告警与评审清单）

## 命名约束（强制）

- 用户主体字段一律 `uid`，禁 `user_id`
- 无软删除：禁 `deleted_at` / `is_deleted`；禁 `created_by` / `updated_by`
- 时间：`create_time` / `update_time`（`BIGINT UNSIGNED` 毫秒，必备），其余业务时间 `{动作}_time` 毫秒，`created_at` / `updated_at` 仅库维护
- 主键 UUIDv7 `BINARY(16)`，禁自增
- 写库 + 发消息必须 Outbox
- 金额必带业务前缀（`order_amount`/`fee_amount`），禁裸 `amount`
- 业务版本字段必须前缀（`rule_version`/`config_version`）

## 输出要求

- 文档第一行下加注释：`<!-- [skill: go-team-standards · 技术方案] $ARGUMENTS -->`
- 末尾单独一行：🌟
