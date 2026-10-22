// Demo: GORM + MySQL 仓储 — internal/data/order_repo.go
// 对应 database.md：ctx + 超时、无软删除（物理删除）、游标分页、乐观锁、自增 id + UUIDv7 oid、create_time/update_time 毫秒
package data

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	apperr "gin-go-common-lib/errors"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	queryTimeout = 3 * time.Second // 所有 DB 调用必带超时（铁律 13）
	maxOIDRetry  = 3               // oid 撞号最多重新生成 3 次（id-generation.md §4）
	oidIndexName = "uk_orders_oid"
)

// OrderPO 对应 trade.orders。字段 / 类型严格对照 database.md §4.10。
//
//	ID  自增主键：库内物理主键，只在本包内部使用（游标 tie-breaker），不得进入 DTO / 事件 / 缓存 Key，不被其他表引用。
//	OID 业务 ID：UUIDv7 小写 36 位，应用层生成；对外、跨表引用（{entity}_oid）一律用它。
type OrderPO struct {
	ID            uint64          `gorm:"column:id;primaryKey;autoIncrement"`
	OID           string          `gorm:"column:oid"`
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

// NewOID 生成业务 ID：UUIDv7，36 位小写。真实项目走公共库 util/idgen（目前该库只有 UUIDv4，需先补 UUIDv7）。
func NewOID() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("gen oid: %w", err)
	}
	return u.String(), nil // google/uuid 的 String() 恒为小写
}

// NormalizeOID 校验并归一为小写 36 位。库默认 utf8mb4_0900_bin 区分大小写：
// 大写 oid 既查不到小写行，还能绕过唯一索引写入重复行，所以进出库前必须归一。
func NormalizeOID(s string) (string, error) {
	u, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("invalid oid %q: %w", s, err)
	}
	return u.String(), nil
}

// OrderRepo 订单仓储。接口定义在 biz 包，这里是实现。
type OrderRepo struct{ db *gorm.DB }

// NewOrderRepo 构造函数注入。
func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{db: db} }

// Create 插入订单。oid 为空则由本函数生成。
//   - 命中 uk_orders_oid（oid 撞号，概率极低）：重新生成 oid 重试，最多 maxOIDRetry 次；
//   - 命中其他唯一键（如 client_order_id 幂等键）：返回 ErrConflict，由 biz 按幂等处理，不得换 oid 重试。
//
// 成功后业务只需要 o.OID；o.ID 是 GORM 回填的自增主键，不要往外传。
func (r *OrderRepo) Create(ctx context.Context, o *OrderPO) error {
	for attempt := 0; attempt < maxOIDRetry; attempt++ {
		if o.OID == "" || attempt > 0 {
			oid, err := NewOID()
			if err != nil {
				return err
			}
			o.OID = oid
		} else {
			oid, err := NormalizeOID(o.OID)
			if err != nil {
				return err
			}
			o.OID = oid
		}
		o.ID = 0
		err := r.insert(ctx, o)
		if err == nil {
			return nil
		}
		key, dup := duplicateKey(err)
		if !dup {
			return fmt.Errorf("create order: %w", err)
		}
		if !strings.HasSuffix(key, oidIndexName) {
			return apperr.ErrConflict("duplicate order") // 业务唯一键冲突 = 幂等重复
		}
		// 命中 oid 唯一索引：进入下一轮，重新生成
	}
	return fmt.Errorf("create order: oid collision after %d attempts", maxOIDRetry)
}

func (r *OrderRepo) insert(ctx context.Context, o *OrderPO) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	return r.db.WithContext(ctx).Create(o).Error
}

// GetByOID 按业务 ID 查询；不存在返回 (nil, nil)，由 biz 转成 OrderNotFound。禁止 SELECT *：显式列出字段。
func (r *OrderRepo) GetByOID(ctx context.Context, oid string) (*OrderPO, error) {
	oid, err := NormalizeOID(oid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	var o OrderPO
	err = r.db.WithContext(ctx).
		Select("id", "oid", "platform_id", "uid", "client_order_id", "symbol", "order_side", "order_type", "price", "quantity",
			"filled_qty", "order_status", "version", "create_time", "update_time").
		Where("oid = ?", oid).Take(&o).Error
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
func (r *OrderRepo) UpdateStatus(ctx context.Context, oid string, status uint8, filled decimal.Decimal, version uint32) (bool, error) {
	oid, err := NormalizeOID(oid)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	res := r.db.WithContext(ctx).Model(&OrderPO{}).
		Where("oid = ? AND version = ?", oid, version).
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

// ListByCursor 游标分页：以 create_time 为主、自增 id 仅作同毫秒 tie-breaker，禁止 OFFSET 大翻页。
// 首次查询传 cursorTime=0。cursorID 只在服务内部使用：返回给前端的游标令牌须做不透明编码，不得直接暴露自增 id。
func (r *OrderRepo) ListByCursor(ctx context.Context, platformID, uid string, cursorTime int64, cursorID uint64, limit int) ([]OrderPO, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	q := r.db.WithContext(ctx).
		Select("id", "oid", "symbol", "order_side", "order_type", "price", "quantity", "filled_qty", "order_status", "create_time").
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

// DeleteByOID 物理删除（本项目不使用软删除）。资金类数据删除前须先归档，见 database.md §4.4。
func (r *OrderRepo) DeleteByOID(ctx context.Context, oid string) error {
	oid, err := NormalizeOID(oid)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err := r.db.WithContext(ctx).Where("oid = ?", oid).Delete(&OrderPO{}).Error; err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	return nil
}

// dupKeyRe 取 MySQL 1062 报错信息里的索引名：Duplicate entry 'x' for key 'orders.uk_orders_oid'
var dupKeyRe = regexp.MustCompile(`for key '([^']+)'`)

// duplicateKey 判断是否 MySQL 唯一键冲突（1062），并返回命中的索引名（形如 "orders.uk_orders_oid"）。
func duplicateKey(err error) (string, bool) {
	var me *mysql.MySQLError
	if !errors.As(err, &me) || me.Number != 1062 {
		return "", false
	}
	if m := dupKeyRe.FindStringSubmatch(me.Message); len(m) == 2 {
		return m[1], true
	}
	return "", true
}
