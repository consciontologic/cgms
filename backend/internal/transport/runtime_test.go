package transport

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestAutomaticFailureReasonDoesNotDiscloseErrorContents(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{
		{nil, "none"},
		{fmt.Errorf("private details: %w", context.DeadlineExceeded), "deadline"},
		{fmt.Errorf("private details: %w", context.Canceled), "canceled"},
		{errors.New("private command or database details"), "transition"},
	} {
		if got := automaticFailureReason(tc.err); got != tc.want {
			t.Fatal(got)
		}
	}
}

func TestAutomaticAllowanceWaitIsBoundedAndResumesAtUTCReset(t *testing.T) {
	waits := automaticWaits{}
	now := time.Date(2026, 10, 1, 23, 59, 50, 0, time.UTC)
	waits.block("exhausted", now)
	if !waits.pending("exhausted", now.Add(9*time.Second)) || waits.pending("another-match", now) {
		t.Fatal("allowance wait did not isolate the exhausted match")
	}
	if waits.pending("exhausted", now.Add(10*time.Second)) {
		t.Fatal("next UTC day must resume admission immediately")
	}
	now = now.Add(time.Hour)
	waits.block("exhausted", now)
	if !waits.pending("exhausted", now.Add(29*time.Second)) || waits.pending("exhausted", now.Add(30*time.Second)) {
		t.Fatal("allowance refresh must stay bounded to thirty seconds")
	}
	for i := 0; i < 2000; i++ {
		waits.block(fmt.Sprintf("match-%d", i), now)
	}
	if len(waits) > 1024 || !waits.pending("match-1999", now) {
		t.Fatal("allowance wait cache is not bounded or dropped its newest admission")
	}
}

func TestAutomaticEntitlementWakeIsNonblockingAndCoalesced(t *testing.T) {
	s := &Server{automaticWake: make(chan struct{}, 1)}
	for i := 0; i < 100; i++ {
		s.wakeAutomatic()
	}
	if len(s.automaticWake) != 1 {
		t.Fatal("wake queue must coalesce repeated entitlement changes")
	}
}

func TestAutomaticIdleVersionsAreBoundedAndWakeOnChange(t *testing.T) {
	idle := automaticIdle{}
	if idle.unchanged("idle", 0) {
		t.Fatal("unseen zero-version match must be checked")
	}
	idle.remember("idle", 0)
	if !idle.unchanged("idle", 0) || idle.unchanged("idle", 1) {
		t.Fatal("only the checked version may skip automatic work")
	}
	// If a command races the scan, remembering the scanned version is conservative:
	// the newly committed version remains eligible for another automatic check.
	idle.remember("idle", 1)
	if idle.unchanged("idle", 2) {
		t.Fatal("newer durable state was suppressed")
	}
	for i := 0; i < 2048; i++ {
		idle.remember(fmt.Sprintf("match-%d", i), int64(i))
	}
	if len(idle.versions) != 1024 || idle.unchanged("idle", 1) || idle.unchanged("match-0", 0) || !idle.unchanged("match-2047", 2047) {
		t.Fatal("idle cache must evict old checks at its fixed bound")
	}
	// Updating an existing entry must not consume another eviction slot.
	for i := 0; i < 2048; i++ {
		idle.remember("match-2047", int64(i))
	}
	if len(idle.versions) != 1024 || !idle.unchanged("match-2046", 2046) {
		t.Fatal("rechecking an existing match evicted unrelated idle versions")
	}
}
