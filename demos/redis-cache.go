// Demo: Redis 缓存 / 分布式锁 — internal/data/cache.go
// 对应 redis.md：Key 由 naming.RedisKey 构造且带 TTL（含随机抖动）；Cache-Aside（先写库后删缓存）；
// redis.Nil 是“未命中”不是错误；singleflight 防击穿；锁用 SET NX PX + Lua 校验 token 释放。
package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"gin-go-common-lib/cache/singleflight"
	"gin-go-common-lib/naming"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const (
	redisTimeout = 300 * time.Millisecond // 读写超时（redis.md §9）
	symbolTTL    = 5 * time.Minute
)

// SymbolConfig 交易对配置（示例缓存对象）。
type SymbolConfig struct {
	Symbol        string `json:"symbol"`
	PriceScale    int    `json:"price_scale"`
	MinQuantity   string `json:"min_quantity"` // 金额类一律 string
	IsTradeEnable bool   `json:"is_trade_enable"`
}

// SymbolLoader 回源（查 MySQL）。接口定义在 biz，这里为演示放在一起。
type SymbolLoader func(ctx context.Context, symbol string) (*SymbolConfig, error)

// SymbolCache 交易对配置缓存。
type SymbolCache struct {
	rdb    goredis.UniversalClient
	load   SymbolLoader
	flight singleflight.Group[*SymbolConfig] // 同进程合并回源，防缓存击穿
}

// NewSymbolCache 构造函数注入。
func NewSymbolCache(rdb goredis.UniversalClient, load SymbolLoader) *SymbolCache {
	return &SymbolCache{rdb: rdb, load: load}
}

// symbolKey 构造 trade:symbol:btc-usdt。naming.RedisKey 只接受小写，交易对（BTC-USDT）必须先转小写，否则会 panic。
func symbolKey(symbol string) string {
	return naming.RedisKey("trade", "symbol", strings.ToLower(symbol))
}

// jitter TTL 加 0~20% 随机抖动，防缓存雪崩。
func jitter(base time.Duration) time.Duration {
	return base + time.Duration(rand.Int64N(int64(base)/5))
}

// Get Cache-Aside 读：先 Redis，未命中回源并回填。
func (c *SymbolCache) Get(ctx context.Context, symbol string) (*SymbolConfig, error) {
	key := symbolKey(symbol)
	rctx, cancel := context.WithTimeout(ctx, redisTimeout)
	raw, err := c.rdb.Get(rctx, key).Bytes()
	cancel()
	switch {
	case err == nil:
		var cfg SymbolConfig
		if jerr := json.Unmarshal(raw, &cfg); jerr == nil {
			return &cfg, nil
		}
		// 缓存内容损坏：当作未命中，回源覆盖
	case errors.Is(err, goredis.Nil):
		// 未命中：不是错误，继续回源
	default:
		// Redis 故障：降级直接回源 MySQL（须有限流保护 MySQL），不向调用方返回错误
	}
	return c.flight.Do(key, func() (*SymbolConfig, error) {
		cfg, lerr := c.load(ctx, symbol)
		if lerr != nil {
			return nil, fmt.Errorf("load symbol %s: %w", symbol, lerr)
		}
		if body, merr := json.Marshal(cfg); merr == nil {
			wctx, wcancel := context.WithTimeout(ctx, redisTimeout)
			_ = c.rdb.Set(wctx, key, body, jitter(symbolTTL)).Err() // 回填失败不影响主流程，但要有日志/指标
			wcancel()
		}
		return cfg, nil
	})
}

// Invalidate 写路径：**先写 MySQL，再调用本方法删缓存**。删除失败须重试 / 发 Kafka 事件补删。
func (c *SymbolCache) Invalidate(ctx context.Context, symbol string) error {
	ctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()
	return c.rdb.Del(ctx, symbolKey(symbol)).Err()
}

// ───────── 分布式锁 ─────────
// 锁只用于削峰 / 防重复提交，不是资金正确性的依据；正确性靠 MySQL 乐观锁 / 唯一键（redis.md §5）。

var unlockScript = goredis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0`)

// ErrLockNotAcquired 获取锁失败。
var ErrLockNotAcquired = errors.New("lock not acquired")

// Locker 分布式锁。
type Locker struct{ rdb goredis.UniversalClient }

// NewLocker 构造函数注入。
func NewLocker(rdb goredis.UniversalClient) *Locker { return &Locker{rdb: rdb} }

// Acquire 加锁：key = lock:{业务域}:{资源类型}:{资源ID}。返回释放函数，调用方 defer 调用。
func (l *Locker) Acquire(ctx context.Context, biz, resource, id string, ttl time.Duration) (release func(), err error) {
	key := naming.RedisKey("lock", biz, resource+":"+id) // lock:{业务域}:{资源类型}:{资源ID}
	token := uuid.NewString()                            // 持有者标识：释放前校验，避免误删别人的锁
	ctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()
	ok, err := l.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("acquire lock %s: %w", key, err)
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}
	return func() {
		rctx, rcancel := context.WithTimeout(context.Background(), redisTimeout)
		defer rcancel()
		_ = unlockScript.Run(rctx, l.rdb, []string{key}, token).Err() // Lua：校验 token 后才 DEL
	}, nil
}
