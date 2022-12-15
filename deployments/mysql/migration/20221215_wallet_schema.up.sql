CREATE DATABASE IF NOT EXISTS `development`;

DROP TABLE IF EXISTS `wallets`;
CREATE TABLE `wallets` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '用戶 UUID',
    `amount` int(11) unsigned NOT NULL COMMENT '用戶餘額',
    PRIMARY KEY (`id`),
    UNIQUE KEY (`amount`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用戶資訊';