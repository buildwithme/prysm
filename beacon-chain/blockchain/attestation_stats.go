package blockchain

import (
	"sync"

	"github.com/prysmaticlabs/prysm/v5/consensus-types/primitives"
	"github.com/sirupsen/logrus"
)

// AttestationVerificationStats holds counters for successful and failed attestations.
// It uses a Mutex to ensure concurrency safety for all accesses.
type AttestationVerificationStats struct {
	sync.Mutex
	successfulCount uint64
	failedReasons   map[string]uint64
}

// NewAttestationVerificationStats creates a new stats struct with a fresh map of failed reasons.
func NewAttestationVerificationStats() *AttestationVerificationStats {
	return &AttestationVerificationStats{
		failedReasons: make(map[string]uint64),
	}
}

// IncrementSuccess increments the count of successfully verified attestations.
func (a *AttestationVerificationStats) IncrementSuccess() {
	a.Lock()
	defer a.Unlock()
	a.successfulCount++
}

// IncrementFailure increments the count for a specific failure reason.
func (a *AttestationVerificationStats) IncrementFailure(reason string) {
	a.Lock()
	defer a.Unlock()
	a.failedReasons[reason]++
}

// SnapshotAndReset returns the current stats and then resets them.
// This ensures we have a clean slate for the next epoch.
func (a *AttestationVerificationStats) SnapshotAndReset() (uint64, map[string]uint64) {
	a.Lock()
	defer a.Unlock()

	// Take a snapshot of current counts
	successes := a.successfulCount
	failures := make(map[string]uint64, len(a.failedReasons))
	for k, v := range a.failedReasons {
		failures[k] = v
	}

	// Reset counts
	a.successfulCount = 0
	a.failedReasons = make(map[string]uint64)

	return successes, failures
}

// LogEpochSummaryAndReset logs the current epoch's attestation verification summary
// (success/failure counts) and then resets the stats for the next epoch.
func (a *AttestationVerificationStats) LogEpochSummaryAndReset(epoch primitives.Epoch) {
	// Snapshot current stats and reset them to avoid mixing data from multiple epochs.
	successes, failures := a.SnapshotAndReset()

	// Prepare log fields: current epoch and number of successful verifications.
	fields := logrus.Fields{
		"epoch":                    epoch,
		"successful_verifications": successes,
	}

	// Add each failure reason and its count to the log fields.
	for reason, count := range failures {
		fields["fail_"+reason] = count
	}

	// Log the summarized data for this epoch, helping operators track trends and issues.
	log.WithFields(fields).Info("Attestation verification summary")
}
