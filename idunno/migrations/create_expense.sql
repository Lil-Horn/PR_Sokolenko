CREATE TABLE IF NOT EXISTS expenses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    amount      REAL    NOT NULL,
    description TEXT    NOT NULL,
    date        TEXT    NOT NULL,
    category_id INTEGER REFERENCES categories(id),
    user_id     INTEGER REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_expenses_user ON expenses(user_id);
