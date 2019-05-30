package expobackoff

import (
	"math/rand"
	"time"
)

const EXPO_LIM = 8

func calculate(num_retries int) []time.Duration {
	retryDurations := make([]time.Duration, num_retries)

	for i := 0; i < num_retries; i++ {
		if num_retries < EXPO_LIM {
			retryDurations[i] = time.Duration(1e9*rand.Float64()) + (1<<uint32(i))*time.Second
		} else {
			retryDurations[i] = time.Duration(1e9*rand.Float64()) + (1<<uint32(EXPO_LIM))*time.Second
		}
	}
	return retryDurations
}
