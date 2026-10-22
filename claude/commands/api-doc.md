---
description: 写接口文档（按 api-doc.md 的 REST + 统一响应体规范与模板）
argument-hint: <接口名 + 用途，如 "POST /api/v1/orders 创建订单">
---

# 写接口文档：$ARGUMENTS

## 步骤

1. **必读**：`~/.claude/skills/go-team-standards/references/api-doc.md`（规范 + 接口文档模板 + 完整示例）
2. **必读**：`~/.claude/skills/go-team-standards/references/error-codes.md`
3. 涉及前端对接：另读 `frontend.md`

## 接口文档结构（强制）

```markdown
# POST /api/v1/<path>

> 一句话描述用途

## 通用约定（如未在通用文档定义）
| 项 | 约定 |
|---|---|
| 基础路径 | /api/v1 |
| 认证 | Bearer Token |
| Content-Type | application/json |

## 请求

### URL
`POST /api/v1/<path>`

### Header
| Key | 必需 | 说明 |
|---|---|---|
| Authorization | ✓ | Bearer Token |
| X-Trace-Id | ○ | 客户端可选，服务端会生成并回传 |
| X-Idempotency-Key | 创建 / 资金类写接口必填 | 幂等键（UUID），重试复用同一个 |

### Body
```json
{
  "field_a": "string",
  "order_amount": "99.00"
}
```

| 字段 | 类型 | 必需 | 说明 | 取值范围 |
|---|---|---|---|---|
| field_a | string | ✓ | ... | 1-64 字符 |
| order_amount | string | ✓ | 金额（**string**，不用 number） | > 0，精度 ≤ 8 位 |

> 字段一律 snake_case；金额 / 数量 / 价格 / 主键 / ID 用 string；业务时间用毫秒 int64（`*_time`）。

## 响应

### 成功 (HTTP 200)
```json
{
  "code": "0",
  "msg": "ok",
  "data": { ... },
  "trace_id": "8c41e0a4-77f2-4d88-9a9d-1f1a9c7df222",
  "timestamp": "2026-10-22T08:15:30Z"
}
```

### 错误码（`{SERVICE}_{MODULE}_{NNN}`，HTTP 状态码如实返回）
| code | HTTP | 含义 | 触发条件 |
|---|---|---|---|
| COMMON_COMMON_002 | 400 | 参数校验失败 | field_a 缺失（`data.errors[]` 含字段明细） |
| TRADE_ORDER_001 | 400 | 业务校验失败 | 价格超出允许区间 |
| COMMON_COMMON_301 | 409 | 状态冲突 / 重复请求 | 重复幂等键 |
| COMMON_COMMON_501 | 500 | 内部异常 | 下游异常 |

## 调用示例

```bash
curl -X POST 'https://api.xxx.com/api/v1/<path>' \
  -H 'Authorization: Bearer xxx' \
  -H 'Content-Type: application/json' \
  -d '{"field_a":"...","field_b":100}'
```

## 边界与注意

- 幂等性：是否幂等？依赖什么 key？
- 限流：每秒上限？
- 超时：上游建议设置多久？
- 兼容性：旧版本表现？字段是否可选？
```

## 命名约束

字段名遵守全局统一命名规范（uid / `*_time` 毫秒 / 金额带前缀且为 string）。

## 输出格式

文档头加：
```
<!-- [skill: go-team-standards · 接口文档] $ARGUMENTS -->
```

末尾单独一行：🌟
