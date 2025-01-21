package utils

import (
	"context"
	"time"
)

func SleepContext(ctx context.Context, t time.Duration) {
	sleep, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	<-sleep.Done()
}

func Setint64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func Setfloat64(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

func probeDefault(err error) bool {
	return true
}

func RetryTimes() []int {
	return []int{1000, 3000, 5000}
}

func RetryAction(
	ctx context.Context,
	timers []int,
	callback func(ctx context.Context) error,
	probers ...func(err error) bool,
) error {
	var err error
	if len(probers) == 0 {
		probers = append(probers, probeDefault)
	}

	for {

		select {
		case <-ctx.Done():

			return ctx.Err()

		default:

			err = callback(ctx)

			if err != nil {
				shouldContinue := false

				for i := 0; i < len(probers); i++ {
					prober := probers[i]

					if prober(err) {
						shouldContinue = true
					}
				}

				if shouldContinue && (len(timers) > 0) {
					SleepContext(ctx, time.Duration(timers[0])*time.Millisecond)
					timers = timers[1:]

					continue
				}

				return err
			}

			return nil
		}
	}
}
