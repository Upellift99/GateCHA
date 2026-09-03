package models_test

import (
	"testing"
	"time"

	"github.com/Upellift99/GateCHA/internal/database"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
	"gorm.io/gorm"
)

// legacyAPIKey is the api_keys shape from before HIS enforcement existed. It
// maps to the same table so a migration can be replayed against rows that
// predate the new columns.
type legacyAPIKey struct {
	ID                 int64  `gorm:"primaryKey;autoIncrement"`
	KeyID              string `gorm:"not null;uniqueIndex;size:32"`
	HMACSecret         string `gorm:"not null"`
	Name               string `gorm:"not null;default:''"`
	Domain             string `gorm:"not null;default:''"`
	MaxNumber          int64  `gorm:"not null;default:100000"`
	ExpireSeconds      int    `gorm:"not null;default:300"`
	Algorithm          string `gorm:"not null;default:'SHA-256'"`
	RateLimitPerMin    int    `gorm:"not null;default:0"`
	AdaptiveDifficulty bool   `gorm:"not null;default:false"`
	HISSampling        bool   `gorm:"not null;default:false"`
	Enabled            bool   `gorm:"not null;default:true"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (legacyAPIKey) TableName() string { return "api_keys" }

// assertUpgradeKeepsWorkingThreshold replays the upgrade: build the old table,
// put a key in it, then migrate to the current model. The key must come back
// with a usable threshold rather than a stored 0, which on an enforcing key
// would mean "reject every scored request".
func assertUpgradeKeepsWorkingThreshold(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Migrator().DropTable(&models.APIKey{}); err != nil {
		t.Fatalf("drop api_keys: %v", err)
	}
	if err := db.AutoMigrate(&legacyAPIKey{}); err != nil {
		t.Fatalf("create legacy api_keys: %v", err)
	}
	if err := db.Create(&legacyAPIKey{KeyID: "gk_legacy", HMACSecret: "x", Name: "Existing", Enabled: true}).Error; err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	if err := database.RunMigrations(db, models.All()...); err != nil {
		t.Fatalf("upgrade migration: %v", err)
	}

	key, err := models.GetAPIKeyByKeyID(db, "gk_legacy")
	if err != nil {
		t.Fatalf("reload upgraded key: %v", err)
	}
	if key.Name != "Existing" {
		t.Errorf("name = %q, the row did not survive the migration", key.Name)
	}
	// Enforcement must not switch itself on for a key nobody configured.
	if key.HISEnforce {
		t.Error("his_enforce must default to off on an upgraded row")
	}
	if got := key.SuspectThreshold(); got != models.DefaultHISThreshold {
		t.Errorf("SuspectThreshold() = %v, want %v", got, models.DefaultHISThreshold)
	}
	// The serialised field matters as much as the computed one: the edit form
	// binds to it directly, and a 0 there is a value its own validation refuses
	// to save back, which would leave the key uneditable.
	if key.HISThreshold != models.DefaultHISThreshold {
		t.Errorf("HISThreshold = %v, want the default to reach the API too", key.HISThreshold)
	}

	// And the column itself carries the default, rather than relying on the
	// read-time fallback to paper over a 0.
	var stored float64
	if err := db.Raw("SELECT his_threshold FROM api_keys WHERE key_id = ?", "gk_legacy").Scan(&stored).Error; err != nil {
		t.Fatalf("read stored threshold: %v", err)
	}
	if stored != models.DefaultHISThreshold {
		t.Errorf("stored his_threshold = %v, want %v: the migration default did not land",
			stored, models.DefaultHISThreshold)
	}
}

func TestUpgradeFromPreEnforcementSchema(t *testing.T) {
	assertUpgradeKeepsWorkingThreshold(t, testutil.SetupTestDB(t))
}

func TestMySQL_UpgradeFromPreEnforcementSchema(t *testing.T) {
	assertUpgradeKeepsWorkingThreshold(t, testutil.SetupTestMySQL(t))
}
