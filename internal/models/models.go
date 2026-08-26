package models

// All returns every persisted model, in dependency order, for migration.
//
// The list lived in three places (the binary, the SQLite test helper and the
// MySQL test helper) and had already drifted: the MySQL helper was missing
// DailyCountryStat and HISSample, so those integration tests ran against an
// incomplete schema. Adding a model here now reaches every caller at once.
func All() []any {
	return []any{
		&AdminUser{},
		&APIKey{},
		&ConsumedChallenge{},
		&DailyStat{},
		&DailyCountryStat{},
		&HISSample{},
		&Setting{},
		&MCPToken{},
	}
}
