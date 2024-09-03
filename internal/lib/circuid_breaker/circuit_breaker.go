package circuid_breaker

import (
	"context"
	"errors"
	"sync"
)

type breakerStateName string

var (
	ErrBreakerOpen = errors.New("breaker is open")
)

const (
	breakerOpen     breakerStateName = "open"
	breakerClose    breakerStateName = "close"
	breakerHalfOpen breakerStateName = "half-open"
)

type Counter struct {
	failure            uint64
	success            uint64
	consecutiveSuccess uint64
	consecutiveFailure uint64
}

func (c *Counter) incFailure() {
	c.failure++
	c.consecutiveFailure++
	c.consecutiveSuccess = 0
}

func (c *Counter) incSuccess() {
	c.success++
	c.consecutiveSuccess++
	c.consecutiveSuccess = 0
}

func (c *Counter) reset() {
	c = &Counter{}
}

func (c *Counter) resetFailures() {
	c.failure = 0
	c.consecutiveFailure = 0
}

func (c *Counter) resetSuccess() {
	c.success = 0
	c.consecutiveSuccess = 0
}

type CircuitBreaker struct {
	mx sync.Mutex

	state breakerState

	onChangeStateHook func(from breakerState, to breakerState)

	counter           *Counter
	failureRateToOpen uint64
}

func NewCircuitBreaker(failureRate uint64) *CircuitBreaker {
	return &CircuitBreaker{
		state:             breakerClose,
		counter:           &Counter{},
		failureRateToOpen: failureRate,
	}
}

type BreakerHandleFn func() (any, error)

func (cb *CircuitBreaker) Handle(ctx context.Context, fn BreakerHandleFn) error {

}

func (cb *CircuitBreaker) Done() error {

}

func (cb *CircuitBreaker) setState(state breakerState) {
	cb.mx.Lock()
	defer cb.mx.Unlock()

	if cb.state != nil {
		cb.state.onExit(cb)
	}

	from := cb.state
	cb.state = state
	cb.state.onEntry(cb)

	cb.handleOnStateChange(from, cb.state)
}

func (cb *CircuitBreaker) handleOnStateChange(from breakerState, to breakerState) {
	if from == nil || to == nil {
		return
	}

	cb.onChangeStateHook(from, to)
}
