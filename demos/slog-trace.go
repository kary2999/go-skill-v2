// Demo: 结构化日志 + trace_id — 公共库 logging（基于 slog，JSON 输出）
// 对应 naming-logging.md / go-style.md §7：必须传 ctx、字段 snake_case、禁止记录敏感信息与完整请求体。
package service

import (
	"context"

	"gin-go-common-lib/logging"
)

// Example 演示日志器的标准用法。
func Example(ctx context.Context) {
	log := logging.New("trade-service") // 服务名进入日志字段 service

	// trace_id 由 middleware.TraceID 注入 ctx，日志器会自动带出；非 HTTP 入口（Kafka 消费、定时任务）自己注入：
	ctx = logging.WithTraceID(ctx, "4bf92f3577b34da6a3ce929d0e0e4736")

	// ✅ 键值对风格，key 一律 snake_case；金额用字符串；日志里带业务定位字段
	log.Info(ctx, "order created",
		"order_id", "0194f3a2-7c1e-7b52-9d0a-3f6c2e8a1b45",
		"uid", "usr_88001",
		"order_amount", "99.00",
		"cost_ms", 12,
	)

	// ✅ 错误日志带 err，不吞错误
	log.Error(ctx, "freeze balance failed", "uid", "usr_88001", "err", context.DeadlineExceeded)

	// ❌ 禁止：fmt.Println / log.Println / 字符串拼接
	// ❌ 禁止：记录 Token、密码、完整证件号 / 卡号、完整请求 / 响应体、Kafka 消息体
	// ❌ 禁止：把 uid / order_id / trace_id 当 Prometheus label（基数爆炸）
}
