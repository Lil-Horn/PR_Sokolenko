-- Выполняется из Go ТОЛЬКО если в старой БД (этапы 3-4) ещё нет колонки user_id.
-- (SQLite не умеет ADD COLUMN IF NOT EXISTS, поэтому наличие колонки проверяется в коде.)
ALTER TABLE expenses ADD COLUMN user_id INTEGER REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_expenses_user ON expenses(user_id);
