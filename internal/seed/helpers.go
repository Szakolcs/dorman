package seed

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const batchSize = 500

const seedPassword = "password"

func batchCreate[T any](db *gorm.DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	return db.CreateInBatches(rows, batchSize).Error
}

func randTime(rng *rand.Rand, start, end time.Time) time.Time {
	if !end.After(start) {
		return start
	}
	delta := end.Sub(start)
	return start.Add(time.Duration(rng.Int64N(int64(delta))))
}

func pick[T any](rng *rand.Rand, items []T) T {
	return items[rng.IntN(len(items))]
}

func pickPtr[T any](rng *rand.Rand, items []T) *T {
	if len(items) == 0 {
		return nil
	}
	v := items[rng.IntN(len(items))]
	return &v
}

func canonicalPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

func studentCode(i int) string {
	return fmt.Sprintf("STU%06d", i+1)
}

func uniCode(prefix string, i int) string {
	return fmt.Sprintf("%s%04d", prefix, i+1)
}

func email(prefix string, i int) string {
	return fmt.Sprintf("%s%d@dorm.local", prefix, i+1)
}
