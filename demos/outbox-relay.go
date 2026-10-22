// Demo: Outbox relay — internal/job/outbox_relay.go
// 对应 kafka.md §5：业务事务里只写 outbox_events；relay 轮询待发送记录发到 Kafka，成功后标记已发送。
// 多实例并行用 FOR UPDATE SKIP LOCKED；同一 msg_key 的事件必须串行（本 demo 按取出顺序串行发送）。
package job

import (
	"context"
	"database/sql/driver"
	"fmt"
	"time"

	"gin-go-common-lib/logging"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	statusPending = 1 // event_status：1=待发送 2=已发送 3=发送失败
	statusSent    = 2
	statusFailed  = 3
	batchSize     = 100 // 单批 ≤ 100（database.md §7.2）
	maxRetries    = 10
)

// BinaryUUID 把 UUID 存成 BINARY(16)（与 mysql-gorm-repo.go 同一实现，真实项目放公共库）。
// 不能用 []byte 当 GORM 的 ? 参数：GORM 会把它展开成 16 个参数。
type BinaryUUID [16]byte

// Value 写库：16 字节。
func (b BinaryUUID) Value() (driver.Value, error) { return b[:], nil }

// Scan 读库：期望 16 字节。
func (b *BinaryUUID) Scan(src any) error {
	raw, ok := src.([]byte)
	if !ok || len(raw) != 16 {
		return fmt.Errorf("BinaryUUID: want 16 bytes, got %T len=%d", src, len(raw))
	}
	copy(b[:], raw)
	return nil
}

// OutboxEvent 对应 trade.outbox_events。
type OutboxEvent struct {
	ID          BinaryUUID `gorm:"column:id;primaryKey"`
	Topic       string     `gorm:"column:topic"`
	MsgKey      string     `gorm:"column:msg_key"`
	Payload     []byte     `gorm:"column:payload"` // JSON 列，已是完整信封
	EventStatus uint8      `gorm:"column:event_status"`
	RetryCount  uint32     `gorm:"column:retry_count"`
	CreateTime  int64      `gorm:"column:create_time"`
	UpdateTime  int64      `gorm:"column:update_time"`
}

// TableName 库名.表名。
func (OutboxEvent) TableName() string { return "trade.outbox_events" }

// Sender 发送一条消息（由 kafkap.Publisher 适配）。
type Sender interface {
	PublishRaw(ctx context.Context, topic string, key, value []byte) error
}

// Relay 轮询发送器。
type Relay struct {
	db     *gorm.DB
	sender Sender
	log    *logging.Logger
}

// NewRelay 构造函数注入。
func NewRelay(db *gorm.DB, sender Sender, log *logging.Logger) *Relay {
	return &Relay{db: db, sender: sender, log: log}
}

// Run 每 200ms 轮询一次；ctx 取消时退出。
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.flushOnce(ctx); err != nil {
				r.log.Error(ctx, "outbox flush failed", "err", err)
			}
		}
	}
}

// flushOnce 在一个事务里：锁住一批待发送记录 → 逐条发送 → 更新状态。
// 事务期间持有行锁，因此整批处理必须很快（有超时），不要在里面做慢操作。
func (r *Relay) flushOnce(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var events []OutboxEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("event_status = ?", statusPending).
			Order("create_time ASC").Limit(batchSize).Find(&events).Error; err != nil {
			return fmt.Errorf("select outbox: %w", err)
		}
		for _, e := range events {
			sendCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err := r.sender.PublishRaw(sendCtx, e.Topic, []byte(e.MsgKey), e.Payload)
			cancel()

			upd := map[string]any{"update_time": time.Now().UnixMilli()} // 每条 UPDATE 必写 update_time
			if err == nil {
				upd["event_status"] = statusSent
			} else {
				upd["retry_count"] = gorm.Expr("retry_count + 1")
				if e.RetryCount+1 >= maxRetries {
					upd["event_status"] = statusFailed // 超限：标记失败并告警，人工处理
					r.log.Error(ctx, "outbox event failed permanently", "topic", e.Topic, "retry", e.RetryCount+1)
				}
			}
			if uerr := tx.Model(&OutboxEvent{}).Where("id = ?", e.ID).Updates(upd).Error; uerr != nil {
				return fmt.Errorf("update outbox: %w", uerr)
			}
			if err != nil {
				break // 保证同一 msg_key 的顺序：这一条失败就停止本批，下一轮从它继续
			}
		}
		return nil
	})
}
