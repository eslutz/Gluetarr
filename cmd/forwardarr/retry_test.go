package main

import (
	"testing"
	"time"
)

func TestExponentialBackoffDelay(t *testing.T) {
	base := 5 * time.Second

	tests := []struct {
		name      string
		attempt   int
		base      time.Duration
		remaining time.Duration
		want      time.Duration
	}{
		{
			name:      "attempt1",
			attempt:   1,
			base:      base,
			remaining: time.Minute,
			want:      5 * time.Second,
		},
		{
			name:      "attempt2",
			attempt:   2,
			base:      base,
			remaining: time.Minute,
			want:      10 * time.Second,
		},
		{
			name:      "attempt3",
			attempt:   3,
			base:      base,
			remaining: time.Minute,
			want:      20 * time.Second,
		},
		{
			name:      "attempt4",
			attempt:   4,
			base:      base,
			remaining: time.Minute,
			want:      40 * time.Second,
		},
		{
			name:      "capToRemaining",
			attempt:   3,
			base:      base,
			remaining: 12 * time.Second,
			want:      12 * time.Second,
		},
		{
			name:      "remainingSmallerThanBase",
			attempt:   1,
			base:      base,
			remaining: 3 * time.Second,
			want:      3 * time.Second,
		},
		{
			name:      "invalidAttempt",
			attempt:   0,
			base:      base,
			remaining: time.Minute,
			want:      0,
		},
		{
			name:      "invalidBase",
			attempt:   1,
			base:      0,
			remaining: time.Minute,
			want:      0,
		},
		{
			name:      "invalidRemaining",
			attempt:   1,
			base:      base,
			remaining: 0,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exponentialBackoffDelay(tt.attempt, tt.base, tt.remaining)
			if got != tt.want {
				t.Errorf("exponentialBackoffDelay() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCalculateMaxAttempts(t *testing.T) {
	base := 5 * time.Second

	tests := []struct {
		name    string
		base    time.Duration
		timeout time.Duration
		want    int
	}{
		{
			name:    "baseTimeout120",
			base:    base,
			timeout: 120 * time.Second,
			want:    5,
		},
		{
			name:    "shortTimeout",
			base:    base,
			timeout: 3 * time.Second,
			want:    1,
		},
		{
			name:    "twoAttempts",
			base:    10 * time.Second,
			timeout: 15 * time.Second,
			want:    2,
		},
		{
			name:    "baseTimeout45",
			base:    base,
			timeout: 45 * time.Second,
			want:    4,
		},
		{
			name:    "invalidBase",
			base:    0,
			timeout: time.Minute,
			want:    1,
		},
		{
			name:    "invalidTimeout",
			base:    base,
			timeout: 0,
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMaxAttempts(tt.base, tt.timeout)
			if got != tt.want {
				t.Errorf("calculateMaxAttempts() = %d, want %d", got, tt.want)
			}
		})
	}
}
