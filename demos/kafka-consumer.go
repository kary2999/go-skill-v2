// Demo: Kafka 消费者 — internal/job/fill_consumer.go
// 对应 kafka.md §4：必须幂等（同事务写去重表）、handler 内自己重试 + 死信（不能依赖“返回 error 让 Kafka 重试”）。
//
// ⚠️ 公共库 mq.Consumer 在 handler 返回 error 时只是 continue：不提交该条，也不会重新投递，
// 后续消息提交成功后 offset 会越过它 → 该消息被静默丢弃。所以本 demo 的 handler 对“已处置”的消息一律返回 nil。
package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gin-go-common-lib/logging"
	"gin-go-common-lib/mq"
	"gin-go-common-lib/naming"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxAttempts = 3

var backoffs = []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second}

// envelope 与发布方的 Envelope 一致。真实项目消费方应引用发布方的 contract 包，禁止自己重新定义。
type envelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	PlatformID string          `json:"platform_id"`
	Payload    json.RawMessage `json:"payload"`
}

// errNonRetryable 表示重试也没用（消息非法、业务规则拒绝），直接进死信。
var errNonRetryable = errors.New("non-retryable")

// DLQPublisher 写死信 Topic 的最小接口（由 kafkap.Publisher 之类实现）。
type DLQPublisher interface {
	PublishRaw(ctx context.Context, topic string, key, value []byte) error
}

// FillConsumer 成交回报消费者。
type FillConsumer struct {
	db    *gorm.DB
	dlq   DLQPublisher
	log   *logging.Logger
	env   string
	group string
}

// NewFillConsumer 构造函数注入。
func NewFillConsumer(db *gorm.DB, dlq DLQPublisher, log *logging.Logger, env string) *FillConsumer {
	return &FillConsumer{db: db, dlq: dlq, log: log, env: env, group: naming.ConsumerGroup("trade_service", "fill_consume")}
}

// Run 阻塞消费；ctx 取消时退出。
func (c *FillConsumer) Run(ctx context.Context, brokers []string) error {
	topic := naming.TopicInProject(c.env, "matching", "trade", "matched")
	consumer := mq.NewConsumer(mq.ConsumerConfig{Brokers: brokers, GroupID: c.group, Topic: topic}, c.handle)
	defer consumer.Close()
	return consumer.Start(ctx)
}

// handle 对每条消息：重试 → 仍失败则写死信 → 返回 nil（本条已处置）。
func (c *FillConsumer) handle(ctx context.Context, msg mq.Message) error {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = c.process(ctx, msg)
		if lastErr == nil {
			return nil
		}
		if errors.Is(lastErr, errNonRetryable) {
			break
		}
		select { // 退避期间也要响应 ctx 取消
		case <-ctx.Done():
			return nil
		case <-time.After(backoffs[attempt]):
		}
	}
	c.log.Error(ctx, "consume failed, send to dlq", "topic", msg.Topic, "offset", msg.Offset, "err", lastErr)
	dlqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := c.dlq.PublishRaw(dlqCtx, msg.Topic+"_dlq", msg.Key, msg.Value); err != nil {
		c.log.Error(ctx, "write dlq failed", "topic", msg.Topic, "offset", msg.Offset, "err", err)
	}
	return nil // 无论如何返回 nil：已重试 + 已尝试死信，不能依赖返回 error 触发重投
}

// process 单次处理：同一事务内“写去重表 + 业务变更 + 写 outbox”。
func (c *FillConsumer) process(ctx context.Context, msg mq.Message) error {
	var env envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil || env.EventID == "" {
		return fmt.Errorf("%w: bad envelope: %v", errNonRetryable, err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 去重：唯一键 (consumer_group, event_id)。1062 说明已处理过，视为成功。
		now := time.Now().UnixMilli()
		err := tx.Exec(
			"INSERT INTO trade.consumed_events (id, consumer_group, event_id, create_time, update_time) VALUES (UUID_TO_BIN(?), ?, ?, ?, ?)",
			newID(), c.group, env.EventID, now, now,
		).Error
		if err != nil {
			var me *mysql.MySQLError
			if errors.As(err, &me) && me.Number == 1062 {
				return nil
			}
			return fmt.Errorf("insert consumed_events: %w", err)
		}
		// 2) 业务变更：乐观锁更新订单 filled_qty / order_status（每条 UPDATE 必写 update_time）
		// 3) 同事务写 outbox_events，由 relay 发布 order_status_updated
		return nil
	})
}

// newID 生成 UUIDv7 的 36 位字符串，配合 SQL 里的 UUID_TO_BIN(?) 写入 BINARY(16)。
// 注意：不要把 []byte 直接当 GORM 的 ? 参数传——GORM 会把它展开成 16 个参数，报 1136 列数不匹配（已在 MySQL 8.4 上实测）。
// 真实项目统一走公共库 idgen，禁止业务代码直接调用 uuid.NewV7。
func newID() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(err) // demo 简化；业务代码必须返回 error，禁止 panic
	}
	return id.String()
}
