package economy

import (
	"testing"
	"time"
)

func TestUTCDayAndPassExtension(t *testing.T) {
	now := time.Date(2026, 9, 30, 23, 30, 0, 0, time.UTC)
	if Day(now.In(time.FixedZone("east", 3*3600))) != "2026-09-30" {
		t.Fatal("local timezone changed allowance")
	}
	if Day(now.Add(time.Hour)) != "2026-10-01" {
		t.Fatal("UTC rollover")
	}
	for _, days := range []int{1, 7} {
		future := now.Add(24 * time.Hour)
		got, e := ExtendPass(now, future, days)
		if e != nil || !got.Equal(future.Add(time.Duration(days)*24*time.Hour)) {
			t.Fatal("stacking", got, e)
		}
		got, e = ExtendPass(now, now.Add(-time.Hour), days)
		if e != nil || !got.Equal(now.Add(time.Duration(days)*24*time.Hour)) {
			t.Fatal("expired pass", got, e)
		}
	}
	for _, days := range []int{-1, 0, 2, 3, 4, 5, 6, 8, 30, 365} {
		if _, e := ExtendPass(now, now, days); e == nil {
			t.Fatal("invalid duration")
		}
	}
}

func TestPolicyDoesNotInventActivePrices(t *testing.T) {
	if (Policy{}).Validate() == nil {
		t.Fatal("missing owner prices activated")
	}
	if (Policy{Reward: 10, PassDayPrice: 30}).Validate() != nil {
		t.Fatal("fixture values")
	}
	for _, p := range []Policy{{Reward: -1, PassDayPrice: 30}, {Reward: 10, PassDayPrice: 0}, {Reward: 1, PassDayPrice: 1 << 62}} {
		if p.Validate() == nil {
			t.Fatal("unsafe policy")
		}
	}
}
