package database

import "database/sql"

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS admin_users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE DEFAULT 'admin',
    password_hash TEXT    NOT NULL,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS api_keys (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id          TEXT    NOT NULL UNIQUE,
    hmac_secret     TEXT    NOT NULL,
    name            TEXT    NOT NULL DEFAULT '',
    domain          TEXT    NOT NULL DEFAULT '',
    max_number      INTEGER NOT NULL DEFAULT 100000,
    expire_seconds  INTEGER NOT NULL DEFAULT 300,
    algorithm       TEXT    NOT NULL DEFAULT 'SHA-256',
    enabled         INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_id ON api_keys(key_id);

CREATE TABLE IF NOT EXISTS consumed_challenges (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    challenge     TEXT    NOT NULL UNIQUE,
    api_key_id    INTEGER NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    expires_at    TEXT    NOT NULL,
    consumed_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_consumed_challenges_challenge ON consumed_challenges(challenge);
CREATE INDEX IF NOT EXISTS idx_consumed_challenges_expires ON consumed_challenges(expires_at);

CREATE TABLE IF NOT EXISTS daily_stats (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_id          INTEGER NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    date                TEXT    NOT NULL,
    challenges_issued   INTEGER NOT NULL DEFAULT 0,
    verifications_ok    INTEGER NOT NULL DEFAULT 0,
    verifications_fail  INTEGER NOT NULL DEFAULT 0,
    UNIQUE(api_key_id, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_stats_key_date ON daily_stats(api_key_id, date);

CREATE TABLE IF NOT EXISTS settings (
    key        TEXT NOT NULL PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`
