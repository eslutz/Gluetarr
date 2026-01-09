package main

import "time"

func exponentialBackoffDelay(attempt int, base, remaining time.Duration) time.Duration {
	if attempt < 1 || base <= 0 || remaining <= 0 {
		return 0
	}

	delay := base
	for i := 1; i < attempt; i++ {
		if delay >= remaining {
			return remaining
		}
		if delay > remaining/2 {
			return remaining
		}
		delay *= 2
	}

	if delay > remaining {
		return remaining
	}
	return delay
}

func calculateMaxAttempts(baseDelay, timeout time.Duration) int {
	if baseDelay <= 0 || timeout <= 0 {
		return 1
	}

	attempts := 1
	remaining := timeout
	for {
		sleep := exponentialBackoffDelay(attempts, baseDelay, remaining)
		if sleep <= 0 || sleep >= remaining {
			return attempts
		}
		remaining -= sleep
		attempts++
	}
}
