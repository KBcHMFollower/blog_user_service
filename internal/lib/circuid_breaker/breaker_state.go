package circuid_breaker

import (
	"github.com/benbjohnson/clock"
	"time"
)

type breakerState interface {
	State() BreakerStateName
	ready() bool
	onEntry(cb *CircuitBreaker)
	onSuccess(cb *CircuitBreaker)
	onFailure(cb *CircuitBreaker)
	onExit(cb *CircuitBreaker)
}

type closedBreakerState struct {
	ticker *clock.Ticker
}

func (cbs *closedBreakerState) State() BreakerStateName { return breakerClose }
func (cbs *closedBreakerState) ready() bool             { return true }
func (cbs *closedBreakerState) onEntry(cb *CircuitBreaker) {
	cb.counter.reset()
	if cb.openConditions.timeInterval == time.Duration(0) {
		return
	}

	cbs.ticker = cb.clock.Ticker(cb.openConditions.timeInterval)
	go func() {
		for {
			select {
			case <-cbs.ticker.C:
				cb.checkOpen()
			default:
				cbs.ticker.Stop()
				return
			}
		}
	}()
}
func (cbs *closedBreakerState) onSuccess(cb *CircuitBreaker) {}
func (cbs *closedBreakerState) onFailure(cb *CircuitBreaker) {
	if cb.checkConsecutiveFailuresCondition() {
		cb.SetState(&openBreakerState{})
	}
}
func (cbs *closedBreakerState) onExit(cb *CircuitBreaker) {
	if cbs.ticker != nil {
		cbs.ticker.Stop()
	}
}

type openBreakerState struct {
	timer *clock.Timer
}

func (obs *openBreakerState) State() BreakerStateName { return breakerOpen }
func (obs *openBreakerState) onEntry(cb *CircuitBreaker) {
	cb.counter.reset()

	obs.timer = cb.clock.AfterFunc(cb.openTime, func() {
		cb.SetState(&halfOpenBreakerState{})
	})
}
func (obs *openBreakerState) onSuccess(cb *CircuitBreaker) {}
func (obs *openBreakerState) onFailure(cb *CircuitBreaker) {}
func (obs *openBreakerState) onExit(cb *CircuitBreaker) {
	if obs.timer != nil {
		obs.timer.Stop()
	}
}
func (obs *openBreakerState) ready() bool { return false }

type halfOpenBreakerState struct {
	timer *clock.Timer
}

func (hobs *halfOpenBreakerState) State() BreakerStateName { return breakerHalfOpen }
func (hobs *halfOpenBreakerState) onEntry(cb *CircuitBreaker) {
	cb.counter.reset()
	if cb.closeConditions.duration == time.Duration(0) {
		return
	}

	hobs.timer = cb.clock.AfterFunc(cb.closeConditions.duration, func() {
		if cb.checkClose() {
			cb.SetState(&closedBreakerState{})
			return
		}

		cb.SetState(&halfOpenBreakerState{})
	})
}
func (hobs *halfOpenBreakerState) onSuccess(cb *CircuitBreaker) {
	if cb.checkConsecutiveSuccessCondition() {
		cb.SetState(&closedBreakerState{})
	}
}
func (hobs *halfOpenBreakerState) onFailure(cb *CircuitBreaker) {}
func (hobs *halfOpenBreakerState) onExit(cb *CircuitBreaker) {
	if hobs.timer != nil {
		hobs.timer.Stop()
	}
}
func (hobs *halfOpenBreakerState) ready() bool { return true }
