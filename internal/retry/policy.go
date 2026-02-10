package retry

import (
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

type Policy struct {
	MaxAttempts  int
	DelaySeconds int
	BackoffBase  int
	MaxDelaySec  int
	RetryOnCodes []int
}

func FromDomain(p *domain.RetryPolicy) Policy {
	if p == nil {
		return Policy{MaxAttempts: 3, DelaySeconds: 30, BackoffBase: 2, MaxDelaySec: 3600}
	}
	return Policy{
		MaxAttempts:  p.MaxAttempts,
		DelaySeconds: p.DelaySeconds,
		BackoffBase:  p.BackoffBase,
		MaxDelaySec:  p.MaxDelaySec,
		RetryOnCodes: p.RetryOnCodes,
	}
}

func (p Policy) ShouldRetry(attempt int, exitCode int) bool {
	if attempt >= p.MaxAttempts {
		return false
	}
	if len(p.RetryOnCodes) == 0 {
		return exitCode != 0
	}
	for _, code := range p.RetryOnCodes {
		if code == exitCode {
			return true
		}
	}
	return false
}

func (p Policy) NextDelay(attempt int) time.Duration {
	if p.BackoffBase <= 1 {
		return time.Duration(p.DelaySeconds) * time.Second
	}
	delay := float64(p.DelaySeconds)
	for i := 1; i < attempt; i++ {
		delay *= float64(p.BackoffBase)
	}
	if int(delay) > p.MaxDelaySec {
		return time.Duration(p.MaxDelaySec) * time.Second
	}
	return time.Duration(delay) * time.Second
}
