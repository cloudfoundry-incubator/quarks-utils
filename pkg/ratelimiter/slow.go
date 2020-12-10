// Package ratelimiter provides custom controller-runtime ratelimiters
package ratelimiter

import (
	"time"

	"golang.org/x/time/rate"

	"k8s.io/client-go/util/workqueue"
)

const (
	baseDelay = 5 * time.Second
	maxDelay  = 1000 * time.Second
	freq      = 10
	burst     = 100
)

// New is a no-arg constructor for a slow rate limiter for a workqueue.  It has
// both overall and per-item rate limiting.  The overall is a token bucket and the per-item is exponential
func New() workqueue.RateLimiter {
	return Custom(baseDelay, maxDelay, freq, burst)
}

// Custom allows to specify all args.
func Custom(
	baseDelay time.Duration,
	maxDelay time.Duration,
	freq int,
	burst int,
) workqueue.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(baseDelay, maxDelay),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(freq), burst)},
	)
}
