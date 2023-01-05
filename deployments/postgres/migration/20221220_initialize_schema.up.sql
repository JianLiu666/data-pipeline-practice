DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    account VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    nickname VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    created_at DATE NOT NULL,
    modified_at DATE NOT NULL
);

COMMENT ON TABLE users IS '用戶資訊';
COMMENT ON COLUMN users.id IS '用戶 UUID';
COMMENT ON COLUMN users.account IS '用戶帳號';
COMMENT ON COLUMN users.password IS '用戶密碼';
COMMENT ON COLUMN users.nickname IS '用戶暱稱';
COMMENT ON COLUMN users.email IS '用戶信箱';
COMMENT ON COLUMN users.created_at IS '註冊日期';
COMMENT ON COLUMN users.modified_at IS '修改日期';

DROP TABLE IF EXISTS wallets;
CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    amount INT NOT NULL,
    created_at DATE NOT NULL,
    modified_at DATE NOT NULL
);

COMMENT ON TABLE wallets IS '錢包資訊';
COMMENT ON COLUMN wallets.id IS '錢包 UUID';
COMMENT ON COLUMN wallets.user_id IS '用戶 UUID';
COMMENT ON COLUMN wallets.amount IS '用戶餘額';
COMMENT ON COLUMN wallets.created_at IS '註冊日期';
COMMENT ON COLUMN wallets.modified_at IS '修改日期';

DROP TABLE IF EXISTS logs;
CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    deposit_user_id INT NOT NULL,
    withdraw_user_id INT NOT NULL,
    amount INT NOT NULL,
    created_at DATE NOT NULL
);

COMMENT ON TABLE logs IS '轉帳記錄';
COMMENT ON COLUMN logs.id IS '日誌 UUID';
COMMENT ON COLUMN logs.deposit_user_id IS '存款用戶 UUID';
COMMENT ON COLUMN logs.withdraw_user_id IS '出款用戶 UUID';
COMMENT ON COLUMN logs.amount IS '轉帳金額';
COMMENT ON COLUMN logs.created_at IS '註冊日期';