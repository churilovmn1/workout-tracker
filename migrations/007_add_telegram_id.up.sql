ALTER TABLE users ADD COLUMN telegram_id BIGINT UNIQUE;
CREATE INDEX idx_users_telegram_id ON users(telegram_id);
