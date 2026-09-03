package models

import (
	"time"

	"gorm.io/gorm"
)

// HISSample is a single stored Human Interaction Signature observation: the raw
// privacy-preserving aggregates produced by the client collector together with
// the score the heuristic assigned. Samples are only written for keys that
// opt in (APIKey.HISSampling) and are pruned after a configurable retention
// window. They exist so HIS enforcement thresholds can be calibrated against
// real traffic before any blocking is enabled.
//
// Like the rest of HIS, this stores no coordinates, timestamps, key contents or
// IP — only counts, distances, durations and timing variance.
type HISSample struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	APIKeyID     int64     `gorm:"not null;index:idx_his_sample_key_created" json:"api_key_id"`
	CreatedAt    time.Time `gorm:"index:idx_his_sample_key_created" json:"created_at"`
	Score        float64   `gorm:"not null" json:"score"`
	BotSuspected bool      `gorm:"not null" json:"bot_suspected"`

	DurationMs         int     `json:"duration_ms"`
	TimeToFirstMs      int     `json:"time_to_first_ms"`
	PointerEvents      int     `json:"pointer_events"`
	PointerDistance    float64 `json:"pointer_distance"`
	Scrolls            int     `json:"scrolls"`
	Touches            int     `json:"touches"`
	Keydowns           int     `json:"keydowns"`
	KeyIntervalStdevMs float64 `json:"key_interval_stdev_ms"`
}

// CreateHISSample persists one raw HIS observation.
func CreateHISSample(db *gorm.DB, s *HISSample) error {
	return db.Create(s).Error
}

// PruneHISSamples deletes samples recorded before cutoff, returning the count
// removed. Called periodically by the cleanup worker to bound storage.
func PruneHISSamples(db *gorm.DB, cutoff time.Time) (int64, error) {
	result := db.Where("created_at < ?", cutoff.UTC()).Delete(&HISSample{})
	return result.RowsAffected, result.Error
}

// HISScoreBucket is one column of the score histogram: scores in [lo, hi).
type HISScoreBucket struct {
	Lo    float64 `json:"lo"`
	Hi    float64 `json:"hi"`
	Count int     `json:"count"`
}

// HISCalibration summarizes stored samples to help tune the scoring heuristic
// and pick an enforcement threshold. The histogram shows how scores distribute
// relative to the current suspect threshold; the signal averages explain what
// is driving them.
type HISCalibration struct {
	Samples        int              `json:"samples"`
	Suspected      int              `json:"suspected"`
	Threshold      float64          `json:"threshold"`
	ScoreHistogram []HISScoreBucket `json:"score_histogram"`
	AvgDurationMs  float64          `json:"avg_duration_ms"`
	AvgPointer     float64          `json:"avg_pointer_events"`
	NoMotionPct    float64          `json:"no_motion_pct"`
}

// GetHISCalibration aggregates stored samples over the last `days`. When
// apiKeyID is nil it spans every key; otherwise it is scoped to that key.
func GetHISCalibration(db *gorm.DB, apiKeyID *int64, days int, threshold float64) (*HISCalibration, error) {
	since := time.Now().UTC().AddDate(0, 0, -days)
	scope := func(q *gorm.DB) *gorm.DB {
		q = q.Where("created_at >= ?", since)
		if apiKeyID != nil {
			q = q.Where("api_key_id = ?", *apiKeyID)
		}
		return q
	}

	cal := &HISCalibration{Threshold: threshold, ScoreHistogram: make([]HISScoreBucket, 10)}
	for i := range cal.ScoreHistogram {
		cal.ScoreHistogram[i] = HISScoreBucket{Lo: float64(i) / 10, Hi: float64(i+1) / 10}
	}

	// Aggregate counts + signal averages in one pass.
	var agg struct {
		Samples       int64
		Suspected     int64
		AvgDurationMs float64
		AvgPointer    float64
		NoMotion      int64
	}
	row := scope(db.Model(&HISSample{})).
		Select("COUNT(*) AS samples, " +
			"COALESCE(SUM(CASE WHEN bot_suspected THEN 1 ELSE 0 END), 0) AS suspected, " +
			"COALESCE(AVG(duration_ms), 0) AS avg_duration_ms, " +
			"COALESCE(AVG(pointer_events), 0) AS avg_pointer, " +
			"COALESCE(SUM(CASE WHEN pointer_events + scrolls + touches = 0 THEN 1 ELSE 0 END), 0) AS no_motion").
		Row()
	if err := row.Scan(&agg.Samples, &agg.Suspected, &agg.AvgDurationMs, &agg.AvgPointer, &agg.NoMotion); err != nil {
		return nil, err
	}
	cal.Samples = int(agg.Samples)
	cal.Suspected = int(agg.Suspected)
	cal.AvgDurationMs = agg.AvgDurationMs
	cal.AvgPointer = agg.AvgPointer
	if agg.Samples > 0 {
		cal.NoMotionPct = float64(agg.NoMotion) / float64(agg.Samples) * 100
	}

	// Score histogram: bucket by floor(score*10), clamped to the last bucket.
	//
	// FLOOR rather than a CAST: the type names CAST accepts are dialect
	// specific, and `CAST(x AS INTEGER)` is a syntax error on MySQL, which
	// wants `SIGNED`. That made the whole endpoint 500 on every MySQL-backed
	// instance. FLOOR is standard and behaves identically on both. It returns
	// a float on SQLite and an integer on MySQL, so the bucket is scanned as a
	// float and truncated here rather than by the driver.
	type bucketRow struct {
		Bucket float64
		Count  int
	}
	var buckets []bucketRow
	if err := scope(db.Model(&HISSample{})).
		Select("FLOOR(score * 10) AS bucket, COUNT(*) AS count").
		Group("bucket").
		Scan(&buckets).Error; err != nil {
		return nil, err
	}
	for _, b := range buckets {
		idx := int(b.Bucket)
		if idx < 0 {
			idx = 0
		}
		if idx > 9 {
			idx = 9
		}
		cal.ScoreHistogram[idx].Count += b.Count
	}

	return cal, nil
}
