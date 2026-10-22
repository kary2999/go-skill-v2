// Demo: Kafka 发布事件 — internal/data/event_publisher.go
// 对应 kafka.md：Topic 由 naming 包构造；统一消息信封；PublishWithHeaders 带 trace_id / platform_id / idempotency_key；Publish 必带超时。
// 注意：业务代码不要在“提交 MySQL 之后”直接调用本发布器——会丢消息。
// 业务事务内只写 outbox_events（见 outbox-relay.go），由 relay 调用本发布器发送。
package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gin-go-common-lib/logging"
	"gin-go-common-lib/mq"
	"gin-go-common-lib/naming"
)

// Envelope 统一消息信封（kafka.md §2.1）。字段 snake_case，时间为毫秒时间戳。
type Envelope struct {
	EventID       string          `json:"event_id"`       // UUIDv7，消费端幂等依据
	EventType     string          `json:"event_type"`     // 如 order.pay_status_updated
	SchemaVersion int             `json:"schema_version"` // 载荷结构版本
	OccurredTime  int64           `json:"occurred_time"`  // 毫秒时间戳 UTC
	PlatformID    string          `json:"platform_id"`
	Payload       json.RawMessage `json:"payload"`
}

// Publisher 事件发布器。
type Publisher struct {
	producer *mq.Producer
	env      string // dev / test / staging / prod
	log      *logging.Logger
}

// NewPublisher 构造函数注入。生产环境 Topic 由运维预建，不依赖 Producer 自动建 Topic。
func NewPublisher(brokers []string, env string, log *logging.Logger) *Publisher {
	naming.MustValidateKafkaEnv(env) // 启动期校验环境名，不合规直接 panic（仅允许在启动阶段）
	return &Publisher{
		producer: mq.NewProducer(mq.ProducerConfig{Brokers: brokers, ClientID: "trade-service"}),
		env:      env,
		log:      log,
	}
}

// OrderStatusTopic 构造订单状态事件 Topic：{env}_{service}_{biz}_{action}，一律走 naming。
func (p *Publisher) OrderStatusTopic() string {
	return naming.TopicInProject(p.env, "trade", "order_status", "updated")
}

// Publish 发送一条事件。key 用业务主键（如 order_id），保证同一订单进同一分区、有序。
func (p *Publisher) Publish(ctx context.Context, topic, key string, env Envelope, traceID string) error {
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second) // Publish 必带超时（铁律 13）
	defer cancel()
	headers := map[string]string{
		"trace_id":        traceID,
		"platform_id":     env.PlatformID,
		"idempotency_key": env.EventID,
	}
	if err := p.producer.PublishWithHeaders(ctx, topic, []byte(key), body, headers); err != nil {
		// 日志不记录消息体（可能含敏感数据），只记录定位信息
		p.log.Error(ctx, "publish event failed", "topic", topic, "event_id", env.EventID, "err", err)
		return fmt.Errorf("publish %s: %w", topic, err)
	}
	return nil
}

// PublishRaw 原样转发一条消息（用于写死信 Topic：{原 topic}_dlq）。
func (p *Publisher) PublishRaw(ctx context.Context, topic string, key, value []byte) error {
	return p.producer.Publish(ctx, topic, key, value)
}

// Close 优雅退出时关闭。
func (p *Publisher) Close() error { return p.producer.Close() }
