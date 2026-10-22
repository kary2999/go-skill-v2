//go:build skill_test_only

// skill-test-case: 违反铁律 #2 —— 错误未走 AppError + errcode
// 期望 AI 指出：业务错误必须返回 *errors.AppError（internal/errcode 定义），否则日志无法按 code
// 聚合、前端无法国际化。给出 errcode 定义 + response.FailErr 用法。

package skilltest

import (
	"errors"
	"fmt"
)

// ❌ 铁律 #2：裸 errors.New 没有 code，监控聚合失败
var ErrOrderNotFound = errors.New("order not found")
var ErrPaymentFailed = errors.New("payment failed: downstream rejected")

func GetOrder(id int64) error {
	if id <= 0 {
		return errors.New("invalid id") // ❌ 没错误码
	}
	// 模拟查库
	return fmt.Errorf("db error: %w", ErrOrderNotFound) // 包装了但里面还是裸字符串
}
