package boxpacker3

import (
	"context"
	"runtime"
	"time"
)

type Budget struct {
	Timeout time.Duration

	Deadline time.Time

	MaxWorkers int
}

func (b Budget) workers() int {
	if b.MaxWorkers > 0 {
		return b.MaxWorkers
	}

	return max(runtime.GOMAXPROCS(0), 1)
}

func (b Budget) deadline(now time.Time) (time.Time, bool) {
	byTimeout := time.Time{}
	if b.Timeout > 0 {
		byTimeout = now.Add(b.Timeout)
	}

	switch {
	case !byTimeout.IsZero() && !b.Deadline.IsZero():
		return earliest(byTimeout, b.Deadline), true
	case !byTimeout.IsZero():
		return byTimeout, true
	case !b.Deadline.IsZero():
		return b.Deadline, true
	}

	return time.Time{}, false
}

func earliest(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}

func (b Budget) apply(ctx context.Context) (context.Context, context.CancelFunc) {
	deadline, limited := b.deadline(time.Now())
	if !limited {
		return context.WithCancel(ctx)
	}

	return context.WithDeadline(ctx, deadline)
}

type Observer interface {
	Candidate(report CandidateReport)

	Improved(result *Result)
}

type ObserverFuncs struct {
	OnCandidate func(report CandidateReport)
	OnImproved  func(result *Result)
}

func (o ObserverFuncs) Candidate(report CandidateReport) {
	if o.OnCandidate != nil {
		o.OnCandidate(report)
	}
}

func (o ObserverFuncs) Improved(result *Result) {
	if o.OnImproved != nil {
		o.OnImproved(result)
	}
}
