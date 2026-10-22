---
description: 设计 MySQL 表（按 database.md + field-naming.md）
argument-hint: <表名 + 业务说明，如 "orders 订单表，含币种、价格、状态">
---

# 设计表：$ARGUMENTS

请按团队规范为我产出**完整的 CREATE TABLE 语句**（MySQL 8.4 / InnoDB）和说明。

## 步骤

1. **必读**：`~/.claude/skills/go-team-standards/references/database.md`
2. **必读**：`~/.claude/skills/go-team-standards/references/field-naming.md`
3. **必读**：`~/.claude/skills/go-team-standards/references/id-generation.md`

## 强制规则

### 必备字段（每张表）

```sql
`id`          BINARY(16)       NOT NULL COMMENT '主键 UUIDv7，应用层生成',
`platform_id` VARCHAR(32)      NOT NULL COMMENT '平台 ID',     -- 多租户
`uid`         VARCHAR(64)      NOT NULL COMMENT '用户 ID',     -- 用户主体（不是 user_id）
-- ... 业务字段 ...
`version`     INT UNSIGNED     NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',   -- 有并发更新时
`create_time` BIGINT UNSIGNED  NOT NULL COMMENT '创建时间，毫秒时间戳（UTC），应用层写入',
`update_time` BIGINT UNSIGNED  NOT NULL COMMENT '更新时间，毫秒时间戳（UTC），每次更新必写',
`created_at`  TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`updated_at`  TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
PRIMARY KEY (`id`)
```

**不要**加：`deleted_at`（无软删除）、`created_by` / `updated_by`、`AUTO_INCREMENT`、`FOREIGN KEY`。

### 命名与类型约束

- 业务时间点：`{动作}_time` `BIGINT UNSIGNED` 毫秒（`paid_time`）；`TIMESTAMP` 仅限 `created_at` / `updated_at`
- 状态 / 类型 / 枚举：`_status` / `_type` `TINYINT UNSIGNED`（`0` 保留未知），禁 `ENUM` / `VARCHAR`；COMMENT 列全取值
- 布尔：`is_` / `has_` / `can_` 开头，`TINYINT UNSIGNED`（`1=是 2=否 0=未知`），禁 `BOOLEAN`
- 金额必带业务前缀：`order_amount` / `fee_amount`，禁裸 `amount`；加密货币 `DECIMAL(38,18)`，法币 `DECIMAL(28,8)`，禁 `FLOAT` / `DOUBLE`
- 数量 `_qty` / `_balance`；价格 `_price` / `_rate`；费率 `DECIMAL(10,8)`
- 建表显式写 `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin`
- 库名 = 业务域（`trade` / `wallet` / `auth` …），禁保留字作表名 / 字段名

### 索引

- 列出所有索引（主键 + 业务索引 + 唯一约束），名称 `uk_<表>_<字段>` / `idx_<表>_<字段>`，≤ 64 字符
- 多租户表索引前缀含 `platform_id`；单表索引 ≤ 8
- 游标分页索引按 `(platform_id, uid, create_time DESC)` 设计
- 软删除不存在，唯一键直接按业务键建

### 容量与分区

- 估算 1 年 / 3 年行数；> 2000 万行须给出按 `create_time` 的分区方案（分区键必须进入主键与所有唯一键）

### 注释

表和每个字段都写内联 `COMMENT`，状态 / 枚举字段列全取值映射。

## 输出格式

```sql
-- [skill: go-team-standards · 数据库设计 · 命名规范] $ARGUMENTS
-- migrations/{6位序号}_{库}_{描述}.sql（goose：-- +goose Up / Down）

CREATE TABLE `xxx`.`yyy` (
    ...
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin COMMENT='...';
```

末尾给一段**字段对照说明**（业务含义、单位、取值范围、是否可空、为什么要这字段），以及容量估算，方便 review。

末尾单独一行：🌟
