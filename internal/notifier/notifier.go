package notifier

import (
	"time"
)

// Payload holds the metadata sent in notifications when a deployment runs.
type Payload struct {
	Project   string
	Status    string // success, failed, rollback
	Duration  time.Duration
	Trigger   string    // github push, manual, etc.
	Timestamp time.Time
	CommitSHA string
	ErrorMsg  string
}
