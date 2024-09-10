package circuid_breaker

import (
	"context"
	"errors"
	"github.com/KBcHMFollower/blog_user_service/internal/lib"
	"github.com/benbjohnson/clock"
	"sync"
	"time"
)

type BreakerStateName string

var (
	ErrBreakerOpen = errors.New("breaker is open")
)

const (
	breakerOpen     BreakerStateName = "open"
	breakerClose    BreakerStateName = "close"
	breakerHalfOpen BreakerStateName = "half-open"
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

type CloseCondition struct {
	duration           time.Duration
	successRate        uint64
	consecutiveSuccess int64
}

type OpenCondition struct {
	timeInterval        time.Duration
	failuresRate        uint64
	consecutiveFailures int64
}

type CBOptions struct {
	clock clock.Clock

	openConditions  OpenCondition
	closeConditions CloseCondition

	openTime        time.Duration
	ignorableErrors []error
}

var defaultCBOptions = CBOptions{
	openTime:        time.Duration(60),
	ignorableErrors: []error{},

	clock: clock.New(),

	openConditions: OpenCondition{
		timeInterval:        time.Duration(60),
		failuresRate:        30,
		consecutiveFailures: 0,
	},

	closeConditions: CloseCondition{
		duration:           time.Duration(60),
		successRate:        90,
		consecutiveSuccess: 5,
	},
}

type CircuitBreaker struct {
	mx sync.Mutex

	state             breakerState
	onChangeStateHook func(from breakerState, to breakerState)

	counter *Counter

	CBOptions
}

func NewCircuitBreaker() *CircuitBreaker {
	cb := &CircuitBreaker{
		counter:   &Counter{},
		CBOptions: defaultCBOptions,
	}
	cb.SetState(&closedBreakerState{})
	return cb
}

func (cb *CircuitBreaker) Configure(conf func(options *CBOptions)) {
	conf(&cb.CBOptions)
}

func (cb *CircuitBreaker) State() BreakerStateName {
	return cb.state.State()
}

func (cb *CircuitBreaker) Success() {
	cb.mx.Lock()
	defer cb.mx.Unlock()
	cb.counter.incSuccess()
	cb.state.onSuccess(cb)
}

func (cb *CircuitBreaker) Failure() {
	cb.mx.Lock()
	defer cb.mx.Unlock()
	cb.counter.incFailure()
	cb.state.onFailure(cb)
}

func (cb *CircuitBreaker) Ready() bool {
	cb.mx.Lock()
	defer cb.mx.Unlock()
	return cb.state.ready()
}

func (cb *CircuitBreaker) Do(ctx context.Context, fn BreakerHandleFn) (any, error) {
	if !cb.Ready() {
		return nil, ErrBreakerOpen
	}

	res, err := fn()
	if err == nil {
		cb.Success()
	}

	if lib.Contains(cb.ignorableErrors, err) {
		cb.Success()
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		cb.Failure()
	}

	cb.Failure()

	return res, err
}

func (cb *CircuitBreaker) SetState(state breakerState) {
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

type BreakerHandleFn func() (any, error)

func (cb *CircuitBreaker) checkOpen() bool {
	return cb.checkFRateCondition()
}

func (cb *CircuitBreaker) checkClose() bool {
	return cb.checkSRateCondition()
}

func (cb *CircuitBreaker) checkFRateCondition() bool {
	counterTotal := cb.counter.failure + cb.counter.success

	if cb.openConditions.failuresRate <= 0 {
		return false
	}

	return cb.counter.failure/counterTotal >= cb.openConditions.failuresRate/100
}

func (cb *CircuitBreaker) checkSRateCondition() bool {
	counterTotal := cb.counter.failure + cb.counter.success

	if cb.closeConditions.successRate <= 0 {
		return false
	}

	return cb.counter.success/counterTotal >= cb.closeConditions.successRate/100
}

func (cb *CircuitBreaker) checkConsecutiveFailuresCondition() bool {
	if cb.openConditions.consecutiveFailures <= 0 {
		return false
	}

	return cb.counter.consecutiveSuccess >= cb.counter.consecutiveFailure
}

func (cb *CircuitBreaker) checkConsecutiveSuccessCondition() bool {
	if cb.closeConditions.consecutiveSuccess <= 0 {
		return false
	}

	return cb.counter.consecutiveSuccess >= cb.counter.consecutiveSuccess
}
