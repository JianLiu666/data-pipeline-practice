ALTER TABLE `wallets`
ADD UNIQUE (amount),
DROP COLUMN user_id,
DROP COLUMN created_at,
DROP COLUMN modified_at;