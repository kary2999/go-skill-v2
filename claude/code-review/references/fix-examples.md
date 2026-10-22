---
title: "code-review 常见违规修复对照表"
version: "1.0.0"
last_modified: "2026-05-01"
---

# 常见违规 → 修复对照（review 时直接套）

## 铁律类

### #1 硬编码密钥

```go
// ❌ 违规
const APIKey = "sk-live-xxx"

// ✅ 修复
import "os"
var APIKey = os.Getenv("API_KEY")
// 或走 config 包
var APIKey = config.GetString("api.key")
```

### #2 errors.New 不走 AppError

```go
// ❌
return errors.New("order not found")

// ✅
return errcode.OrderNotFound() // internal/errcode 定义；handler 用 response.FailErr 输出
```

### #6 金额 float

```go
// ❌
type Wallet struct {
    Balance float64
}

// ✅
import "github.com/shopspring/decimal"
type Wallet struct {
    Balance decimal.Decimal
}
```

### #13 IO 无超时 / 无 ctx

```go
// ❌
resp, err := http.Get("https://api.example.com/")

// ✅
client := httpclient.New(httpclient.Timeout(5*time.Second))
resp, err := client.Get(ctx, "https://api.example.com/")
```

## 命名类

### §1.2 user_id → uid

```sql
-- ❌
CREATE TABLE orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL
);

-- ✅
CREATE TABLE `trade`.`orders` (
    `id` BINARY(16) NOT NULL COMMENT '主键 UUIDv7，应用层生成',
    `uid` VARCHAR(64) NOT NULL COMMENT '用户 ID',
    `platform_id` VARCHAR(32) NOT NULL COMMENT '平台 ID',
    PRIMARY KEY (`id`)
);
```

### §2.2 gmt_create → create_time / created_at

```sql
-- ❌
gmt_create TIMESTAMP,
gmt_modified TIMESTAMP

-- ✅ 毫秒时间戳由应用层写；TIMESTAMP 由库维护，仅运维参考
create_time BIGINT UNSIGNED NOT NULL COMMENT '创建时间，毫秒时间戳（UTC）',
update_time BIGINT UNSIGNED NOT NULL COMMENT '更新时间，毫秒时间戳（UTC），每次更新必写',
created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
```

### §5.2 / database §4.4 不使用软删除

```sql
-- ❌
is_deleted BOOLEAN DEFAULT FALSE      -- 软删除字段
deleted_at TIMESTAMP NULL             -- 同样不允许

-- ✅ 本项目不使用软删除：直接物理 DELETE（资金类先归档到 hist_* / log_*），
--    需要保留状态时用 xxx_status（如 4=已作废）
```

### §4.2 裸 amount → 业务前缀

```sql
-- ❌
amount NUMERIC(28, 8)

-- ✅（按业务）
order_amount NUMERIC(28, 8),     -- 订单金额
fee_amount NUMERIC(28, 8),        -- 手续费
trade_amount NUMERIC(28, 8)       -- 成交金额
```

### §5.1 布尔非 is_ 前缀

```sql
-- ❌
mfa_enabled BOOLEAN,
deposit_enabled BOOLEAN,
auto_renewal BOOLEAN

-- ✅
is_mfa_enabled BOOLEAN,
is_deposit_enabled BOOLEAN,
is_auto_renewal BOOLEAN
```

### §6.3 裸 version 业务语义

```sql
-- ❌
version INT  -- 但实际是规则版本

-- ✅
version INT,           -- 乐观锁（保留）
rule_version INT,      -- 风控规则版本
config_version INT,    -- 配置版本
formula_version INT    -- 收益公式版本
```

## TODO 注释统一格式

```
TODO[skill: code-review · 铁律 #N]: 原 → 推荐
TODO[skill: code-review · 命名 §X.X]: 原 → 推荐
```

例：
```sql
-- TODO[skill: code-review · 命名 §1.2]: user_id BIGINT → uid VARCHAR(64)
user_id BIGINT NOT NULL,

-- TODO[skill: code-review · 命名 §2.2]: gmt_create TIMESTAMP → create_time BIGINT UNSIGNED（毫秒）+ created_at TIMESTAMP（库维护）
gmt_create TIMESTAMP NOT NULL,
```
