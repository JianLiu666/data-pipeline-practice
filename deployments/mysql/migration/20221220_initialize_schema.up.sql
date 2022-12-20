CREATE DATABASE IF NOT EXISTS `development`;

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '用戶 UUID',
    `account` varchar(255) NOT NULL COMMENT '用戶帳號',
    `password` text NOT NULL COMMENT '用戶密碼',
    `nickname` varchar(255) NOT NULL COMMENT '用戶暱稱',
    `email` varchar(255) NOT NULL COMMENT '用戶信箱',
    `created_at` datetime NOT NULL COMMENT '註冊日期',
    `modified_at` datetime NOT NULL COMMENT '修改日期',
    PRIMARY KEY (`id`),
    UNIQUE KEY (`account`),
    UNIQUE KEY (`nickname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用戶資訊';

DROP TABLE IF EXISTS `wallets`;
CREATE TABLE `wallets` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '錢包 UUID',
    `user_id` int(11) unsigned NOT NULL COMMENT '用戶 UUID',
    `amount` int(11) unsigned NOT NULL COMMENT '用戶餘額',
    `created_at` datetime NOT NULL COMMENT '註冊日期',
    `modified_at` datetime NOT NULL COMMENT '修改日期',
    PRIMARY KEY (`id`),
    UNIQUE KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='錢包資訊';

DROP TABLE IF EXISTS `logs`;
CREATE TABLE `logs` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '日誌 UUID',
    `deposit_user_id` int(11) unsigned NOT NULL COMMENT '存款用戶 UUID',
    `withdraw_user_id` int(11) unsigned NOT NULL COMMENT '出款用戶 UUID',
    `amount` int(11) unsigned NOT NULL COMMENT '轉帳金額',
    `created_at` datetime NOT NULL COMMENT '註冊日期',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='轉帳記錄';