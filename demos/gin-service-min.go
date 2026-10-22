// Demo: 最小 Gin 服务骨架 — cmd/trade/main.go
// 使用 gin-go-common-lib：logging / middleware / response / errors / idempotency / storage
// 分层：handler → biz → data（此处为演示合并在一个文件，真实项目按 go-style.md §2 拆到 internal/ 各包）
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apperr "gin-go-common-lib/errors"
	"gin-go-common-lib/idempotency"
	"gin-go-common-lib/logging"
	"gin-go-common-lib/metrics"
	"gin-go-common-lib/middleware"
	"gin-go-common-lib/response"
	"gin-go-common-lib/storage/db"
	rds "gin-go-common-lib/storage/redis"
	"gin-go-common-lib/transport/header"

	"github.com/gin-gonic/gin"
)

// ───────── biz 层：业务规则，不依赖 gin / gorm / redis ─────────

// CreateOrderInput 是 biz 层的入参（与 HTTP 请求结构解耦）。
type CreateOrderInput struct {
	UID      string
	Symbol   string
	Price    string // 金额类一律 string，biz 内用 decimal 运算
	Quantity string
}

// OrderUsecase 下单用例。repo 等依赖通过构造函数注入。
type OrderUsecase struct {
	log *logging.Logger
}

// NewOrderUsecase 构造函数注入，不使用全局变量。
func NewOrderUsecase(log *logging.Logger) *OrderUsecase { return &OrderUsecase{log: log} }

// CreateOrder 返回订单 ID。业务错误统一返回 *errors.AppError（见 apperror.go）。
func (u *OrderUsecase) CreateOrder(ctx context.Context, in CreateOrderInput) (string, error) {
	if in.Symbol == "" {
		return "", apperr.ErrValidation("symbol is required")
	}
	// ... 校验精度 → 冻结资金 → 事务内写 orders + outbox_events → 返回 ID（见 kafka.md §5）
	u.log.Info(ctx, "order created", "uid", in.UID, "symbol", in.Symbol)
	return "0194f3a2-7c1e-7b52-9d0a-3f6c2e8a1b45", nil
}

// ───────── handler 层：只做绑定校验 → 调 biz → 输出 ─────────

// CreateOrderReq JSON 字段一律 snake_case；金额类用 string。
type CreateOrderReq struct {
	Symbol   string `json:"symbol" binding:"required"`
	Price    string `json:"price" binding:"required"`
	Quantity string `json:"quantity" binding:"required"`
}

// OrderHandler HTTP 适配层。
type OrderHandler struct{ uc *OrderUsecase }

// Create POST /api/v1/orders
func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderReq
	if !response.BindJSONOrFail(c, &req) {
		return
	}
	// X-User-Id 由网关鉴权后注入，经 middleware.TraceID 放入 ctx
	uid, ok := header.GetUserID(c.Request.Context())
	if !ok {
		response.FailErr(c, apperr.ErrUnauthorized("missing user"))
		return
	}
	id, err := h.uc.CreateOrder(c.Request.Context(), CreateOrderInput{
		UID: uid, Symbol: req.Symbol, Price: req.Price, Quantity: req.Quantity,
	})
	if err != nil {
		response.FailErr(c, err) // AppError 自动映射 HTTP 状态码与业务码
		return
	}
	response.JSON(c, gin.H{"order_id": id})
}

// ───────── 启动入口：读配置 → 构造依赖 → 注册路由 → 优雅退出 ─────────

func main() {
	log := logging.New("trade-service")
	reg := metrics.NewRegistry()

	// MySQL：公共库已带慢查询日志与连接池（单服务上限 ≤ 50，见 database.md §2.3）。
	// DSN 必带：charset=utf8mb4&parseTime=true&loc=UTC&clientFoundRows=true&timeout=3s&readTimeout=35s&writeTimeout=5s
	gdb, err := db.NewMySQL(db.Config{
		Driver:       "mysql",
		Source:       os.Getenv("MYSQL_DSN"), // 密钥走环境变量，禁止硬编码
		MaxOpenConns: 50,
		MaxIdleConns: 10,
	}, log)
	if err != nil {
		log.Error(context.Background(), "init mysql failed", "err", err)
		os.Exit(1)
	}
	_ = gdb // 传给 data 层的 repo

	rdb, err := rds.New(rds.Config{Mode: "single", Addr: os.Getenv("REDIS_ADDR"), Password: os.Getenv("REDIS_PASSWORD")})
	if err != nil {
		log.Error(context.Background(), "init redis failed", "err", err)
		os.Exit(1)
	}

	r := gin.New()
	r.Use(middleware.Default(log, reg)...) // TraceID → Recovery → Logger → Metrics
	r.GET("/healthz", func(c *gin.Context) { response.JSON(c, gin.H{"status": "ok"}) })
	r.GET(metrics.DefaultPath, gin.WrapH(reg.Handler()))

	h := &OrderHandler{uc: NewOrderUsecase(log)}
	v1 := r.Group("/api/v1")
	// 幂等：X-Idempotency-Key；Redis 异常时放行，最终幂等仍靠 MySQL 唯一键（redis.md §6）
	v1.POST("/orders", idempotency.Middleware(idempotency.NewRedisStore(rdb), 24*time.Hour), h.Create)

	srv := &http.Server{Addr: ":8080", Handler: r, ReadHeaderTimeout: 5 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() { // 受 ctx / Shutdown 管控的 goroutine
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(ctx, "http server exited", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error(shutdownCtx, "http shutdown failed", "err", err)
	}
}
