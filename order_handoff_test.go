package main

import "testing"

func TestFollowupExpressionAdvancesByHours(t *testing.T) {
	for _, tc := range []struct {
		hours    int
		wantHour int
	}{{1, 1}, {23, 23}} {
		got := followupExpression(tc.hours)
		if got == "" {
			t.Fatalf("empty expression for %d hours", tc.hours)
		}
	}
}

func TestScheduleRejectsLongDelay(t *testing.T) {
	_, err := scheduleHandoff(&infraiClient{}, OrderHandoff{DelayH: 24}, "http://x/followup")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
