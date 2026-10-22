// Demo: 业务错误码 — internal/errcode/order.go
// 对应 error-codes.md §5：错误码字面量只允许出现在 errcode 包；其余代码一律引用这里的函数。
// 先在 docs/error-codes.yaml 登记，再在这里定义（同一个 PR）。
package errcode

import "gin-go-common-lib/errors"

// 订单模块（SERVICE=TRADE, MODULE=ORDER）。
// 用「返回新实例的函数」而不是包级 var，避免多处共享同一指针被误改。

// OrderPriceOutOfRange 价格超出允许区间（001-099 业务校验，HTTP 400）。
func OrderPriceOutOfRange() *errors.AppError {
	return &errors.AppError{Code: "TRADE_ORDER_001", Msg: "order price is out of the allowed range", HTTPStatus: 400}
}

// OrderNotFound 订单不存在（200-299 资源不存在，HTTP 404）。
func OrderNotFound() *errors.AppError {
	return &errors.AppError{Code: "TRADE_ORDER_201", Msg: "order not found", HTTPStatus: 404}
}

// OrderStatusConflict 订单状态流转冲突 / 乐观锁冲突（300-399，HTTP 409）。
func OrderStatusConflict() *errors.AppError {
	return &errors.AppError{Code: "TRADE_ORDER_301", Msg: "order status transition conflict", HTTPStatus: 409}
}

// 使用方式（biz 层返回，handler 层统一出口）：
//
//	// biz
//	order, err := uc.repo.GetByID(ctx, id)
//	if err != nil {
//	    if errors.Is(err) { return nil, err }            // 已是 AppError，原样上抛
//	    return nil, fmt.Errorf("get order %s: %w", id, err) // 其余包装保留错误链
//	}
//	if order == nil { return nil, errcode.OrderNotFound() }
//
//	// handler
//	if err != nil { response.FailErr(c, err); return }
//
// 注意：非 AppError 会被 FailErr 当作 COMMON_COMMON_501 返回且不暴露原文，
// 所以原始 error 必须在返回前由调用方记录日志（带 trace_id）。
