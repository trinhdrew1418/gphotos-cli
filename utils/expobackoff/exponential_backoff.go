package expobackoff

import (
	"math/rand"
	"net/http"
	"time"
)

const EXPO_LIM = 16
const NUM_RETRIES = 60

func Calculate(num_retries int) []time.Duration {
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

func RequestUntilSuccess(action func(*http.Request) (*http.Response, error), r *http.Request) (*http.Response, error) {
	resp, err := action(r)

	if err != nil || resp.StatusCode != 200 {
		durations := Calculate(NUM_RETRIES)
		for _, sleepDur := range durations {
			duration := time.Duration(sleepDur)
			time.Sleep(duration)
			resp, err = action(r)

			if err == nil && resp.StatusCode == 200 {
				return resp, err
			}
		}
	}

	return resp, err
}
