package utils

import (
	"context"
	"sync"
	"time"
)

// RunJobsWithContext run each job with a goroutine with context and wait group
func RunJobsWithContext(ctx context.Context, wg *sync.WaitGroup, jobs ...func(context.Context)) {
	for _, job := range jobs {
		wg.Add(1)
		go func(job func(context.Context)) {
			defer wg.Done()
			job(ctx)
		}(job)
	}
}

// RunStateless run an interval jobs without waiting for it to complete each time
func RunStateless(ctx context.Context, interval time.Duration, job func()) {
	ticker := time.NewTicker(interval)
	done := ctx.Done()
	for {
		select {
		case <-ticker.C:
			go job()

		case <-done:
			return
		}
	}
}

// RunStateless run an interval jobs and waiting for it to complete each time
func RunStateful(ctx context.Context, dur time.Duration, job func()) {
	for {
		job()
		if ctx.Err() != nil {
			break
		}
		time.Sleep(dur)
	}
}
