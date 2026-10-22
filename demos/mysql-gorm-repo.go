// Demo: GORM + MySQL 仓储 — internal/data/order_repo.go
// 对应 database.md：ctx + 超时、无软删除（物理删除）、游标分页、乐观锁、UUIDv7 BINARY(16)、create_time/update_time 毫秒
package data

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	apperr "gin-go-common-lib/errors"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const queryTimeout = 3 * time.Second // 所有 DB 调用必带超时（铁律 13）

// BinaryUUID 把 uuid.UUID 存成 BINARY(16)。
// google/uuid 的 Value() 返回 36 位字符串，直接写 BINARY(16) 列会出错或被截断，所以必须包一层。
// 真实项目应放进 gin-go-common-lib，各服务共用。
type BinaryUUID uuid.UUID

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

// String 对外输出标准 36 位字符串。
func (b BinaryUUID) String() string { return uuid.UUID(b).String() }

// OrderPO 对应 trade.orders。字段 / 类型严格对照 database.md §4.10。
type OrderPO struct {
	ID            BinaryUUID      `gorm:"column:id;primaryKey"`
	PlatformID    string          `gorm:"column:platform_id"`
	UID           string          `gorm:"column:uid"`
	ClientOrderID string          `gorm:"column:client_order_id"`
	Symbol        string          `gorm:"column:symbol"`
	OrderSide     uint8           `gorm:"column:order_side"` // 1=买 2=卖
	OrderType     uint8           `gorm:"column:order_type"` // 1=限价 2=市价
	Price         decimal.Decimal `gorm:"column:price;type:decimal(38,18)"`
	Quantity      decimal.Decimal `gorm:"column:quantity;type:decimal(38,18)"`
	FilledQty     decimal.Decimal `gorm:"column:filled_qty;type:decimal(38,18)"`
	OrderStatus   uint8           `gorm:"column:order_status"`
	Version       uint32          `gorm:"column:version"`
	// 毫秒时间戳由应用层写入；GORM 的 autoCreateTime/autoUpdateTime:milli 会自动填，
	// 但手写 SQL / 批量更新不会触发，必须手动写 update_time（见 UpdateStatus）。
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli"`
	UpdateTime int64 `gorm:"column:update_time;autoUpdateTime:milli"`
	// created_at / updated_at 由数据库维护，应用层只读（<-:false），业务代码不得读取。
	CreatedAt time.Time `gorm:"column:created_at;<-:false"`
	UpdatedAt time.Time `gorm:"column:updated_at;<-:false"`
}

// TableName 库名.表名。
func (OrderPO) TableName() string { return "trade.orders" }

// OrderRepo 订单仓储。接口定义在 biz 包，这里是实现。
type OrderRepo struct{ db *gorm.DB }

// NewOrderRepo 构造函数注入。
func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{db: db} }

// Create 插入订单。命中唯一键（1062）返回 ErrConflict，由 biz 决定按幂等处理。
func (r *OrderRepo) Create(ctx context.Context, o *OrderPO) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err := r.db.WithContext(ctx).Create(o).Error; err != nil {
		if isDuplicate(err) {
			return apperr.ErrConflict("duplicate order")
		}
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

// GetByID 按主键查询；不存在返回 (nil, nil)，由 biz 转成 OrderNotFound。禁止 SELECT *：显式列出字段。
func (r *OrderRepo) GetByID(ctx context.Context, id BinaryUUID) (*OrderPO, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	var o OrderPO
	err := r.db.WithContext(ctx).
		Select("id", "platform_id", "uid", "client_order_id", "symbol", "order_side", "order_type", "price", "quantity",
			"filled_qty", "order_status", "version", "create_time", "update_time").
		Where("id = ?", id).Take(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &o, nil
}

// UpdateStatus 乐观锁更新。返回 false 表示版本冲突（需 DSN 开启 clientFoundRows=true，否则“值没变”也会返回 0）。
// 每条 UPDATE 都必须写 update_time（database.md §4.1）。
func (r *OrderRepo) UpdateStatus(ctx context.Context, id BinaryUUID, status uint8, filled decimal.Decimal, version uint32) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	res := r.db.WithContext(ctx).Model(&OrderPO{}).
		Where("id = ? AND version = ?", id, version).
		Updates(map[string]any{
			"order_status": status,
			"filled_qty":   filled,
			"version":      gorm.Expr("version + 1"),
			"update_time":  time.Now().UnixMilli(),
		})
	if res.Error != nil {
		return false, fmt.Errorf("update order status: %w", res.Error)
	}
	return res.RowsAffected == 1, nil
}

// ListByCursor 游标分页：以 create_time 为主、id 仅作同毫秒 tie-breaker，禁止 OFFSET 大翻页。
// 首次查询传 cursorTime=0。
func (r *OrderRepo) ListByCursor(ctx context.Context, platformID, uid string, cursorTime int64, cursorID BinaryUUID, limit int) ([]OrderPO, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	q := r.db.WithContext(ctx).
		Select("id", "symbol", "order_side", "order_type", "price", "quantity", "filled_qty", "order_status", "create_time").
		Where("platform_id = ? AND uid = ?", platformID, uid)
	if cursorTime > 0 {
		q = q.Where("(create_time < ? OR (create_time = ? AND id < ?))", cursorTime, cursorTime, cursorID)
	}
	var list []OrderPO
	if err := q.Order("create_time DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	return list, nil
}

// DeleteByID 物理删除（本项目不使用软删除）。资金类数据删除前须先归档，见 database.md §4.4。
func (r *OrderRepo) DeleteByID(ctx context.Context, id BinaryUUID) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&OrderPO{}).Error; err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	return nil
}

// isDuplicate 判断 MySQL 唯一键冲突（错误号 1062）。
func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}
