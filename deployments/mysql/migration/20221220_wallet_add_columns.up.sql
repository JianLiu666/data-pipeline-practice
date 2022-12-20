TRUNCATE TABLE `wallets`;

ALTER TABLE `wallets`
DROP INDEX amount,
ADD COLUMN user_id int(11) unsigned NOT NULL COMMENT '用戶 UUID' AFTER id,
ADD UNIQUE (user_id),
ADD COLUMN created_at datetime NOT NULL COMMENT '註冊日期',
ADD COLUMN modified_at datetime NOT NULL COMMENT '修改日期';