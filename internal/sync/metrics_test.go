package sync

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsHelpers(t *testing.T) {
	SetCurrentPort(4242)
	if got := testutil.ToFloat64(currentPort); got != 4242 {
		t.Fatalf("currentPort = %v, want 4242", got)
	}

	baselineTotal := testutil.ToFloat64(syncTotal)
	IncrementSyncTotal()
	if got := testutil.ToFloat64(syncTotal); got != baselineTotal+1 {
		t.Fatalf("syncTotal = %v, want %v", got, baselineTotal+1)
	}

	baselineErrors := testutil.ToFloat64(syncErrors)
	IncrementSyncErrors()
	if got := testutil.ToFloat64(syncErrors); got != baselineErrors+1 {
		t.Fatalf("syncErrors = %v, want %v", got, baselineErrors+1)
	}

	UpdateLastSyncTimestamp()
	if got := testutil.ToFloat64(lastSyncTimestamp); got <= float64(time.Now().Add(-1*time.Second).Unix()) {
		t.Fatalf("lastSyncTimestamp not updated, got %v", got)
	}
}
