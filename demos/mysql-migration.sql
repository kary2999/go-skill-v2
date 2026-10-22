-- Demo: goose 迁移模板（MySQL 8.4）— migrations/000001_trade_create_orders.sql
-- 对应 database.md §4.10 / §6：一个文件只改一个库、尽量只放一条 DDL（MySQL DDL 隐式提交，无法回滚半成品）。
-- 文件命名：{6位序号}_{库}_{描述}.sql（goose 单文件含 Up / Down）。
-- 必备字段：id(BINARY(16) UUIDv7) / create_time / update_time(毫秒) / created_at / updated_at(TIMESTAMP，库维护)
-- 禁止：AUTO_INCREMENT、FOREIGN KEY、deleted_at / 软删除、created_by / updated_by、ENUM / BOOLEAN、FLOAT / DOUBLE、触发器 / 存储过程

-- +goose Up
-- +goose StatementBegin
CREATE TABLE `trade`.`orders` (
    `id`               BINARY(16)       NOT NULL COMMENT '主键 UUIDv7，应用层 idgen 生成',

    -- 多租户必备
    `platform_id`      VARCHAR(32)      NOT NULL COMMENT '平台 ID',
    `uid`              VARCHAR(64)      NOT NULL COMMENT '用户 ID',

    -- 业务字段
    `client_order_id`  VARCHAR(64)      NOT NULL COMMENT '客户端订单号，幂等键',
    `symbol`           VARCHAR(40)      NOT NULL COMMENT '交易对 BASE-QUOTE',
    `order_side`       TINYINT UNSIGNED NOT NULL COMMENT '方向：1=买 2=卖',
    `order_type`       TINYINT UNSIGNED NOT NULL COMMENT '类型：1=限价 2=市价',
    `price`            DECIMAL(38,18)   NOT NULL COMMENT '委托价',
    `quantity`         DECIMAL(38,18)   NOT NULL COMMENT '委托数量',
    `filled_qty`       DECIMAL(38,18)   NOT NULL DEFAULT 0 COMMENT '已成交数量',
    `order_status`     TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1=待撮合 2=部分成交 3=完全成交 4=已撤销 5=已拒绝',

    -- 并发控制（乐观锁；裸 version 只用于乐观锁）
    `version`          INT UNSIGNED     NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',

    -- 必备时间字段
    `create_time`      BIGINT UNSIGNED  NOT NULL COMMENT '创建时间，毫秒时间戳（UTC），应用层写入',
    `update_time`      BIGINT UNSIGNED  NOT NULL COMMENT '更新时间，毫秒时间戳（UTC），应用层每次更新必写',
    `created_at`       TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_orders_platform_uid_client_order` (`platform_id`, `uid`, `client_order_id`),
    KEY `idx_orders_platform_uid_ctime` (`platform_id`, `uid`, `create_time` DESC),
    KEY `idx_orders_platform_symbol_status` (`platform_id`, `symbol`, `order_status`),
    CONSTRAINT `ck_orders_positive_qty` CHECK (`quantity` > 0),
    CONSTRAINT `ck_orders_valid_side`   CHECK (`order_side` IN (1, 2)),
    CONSTRAINT `ck_orders_filled_qty`   CHECK (`filled_qty` >= 0)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_0900_bin
  ROW_FORMAT=DYNAMIC
  COMMENT='现货订单表';
-- +goose StatementEnd

-- 大表（> 2000 万行）须按 create_time 分区：分区键必须进入主键与所有唯一键，见 database.md §5.1。
-- 在线加索引示例（不要去掉 ALGORITHM 子句强跑；不支持时改走 gh-ost）：
--   ALTER TABLE `trade`.`orders` ADD INDEX idx_xxx (...), ALGORITHM=INPLACE, LOCK=NONE;

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `trade`.`orders`;
-- +goose StatementEnd
