package circuid_breaker

type breakerState interface {
	State() breakerStateName
	ready() bool
	onEntry(cb *CircuitBreaker)
	onSuccess(cb *CircuitBreaker)
	onFailure(cb *CircuitBreaker)
	onExit(cb *CircuitBreaker)
}

type closedBreakerState struct{}

func (cbs *closedBreakerState) State() breakerStateName      { return breakerClose }
func (cbs *closedBreakerState) ready() bool                  { return true }
func (cbs *closedBreakerState) onEntry(cb *CircuitBreaker)   { cb.counter.reset() }
func (cbs *closedBreakerState) onSuccess(cb *CircuitBreaker) { cb.counter.incSuccess() }
func (cbs *closedBreakerState) onFailure(cb *CircuitBreaker) { cb.counter.incFailure() }
func (cbs *closedBreakerState) onExit(cb *CircuitBreaker)    {}

type openBreakerState struct{}

func (obs *openBreakerState) State() breakerStateName      { return breakerOpen }
func (obs *openBreakerState) onEntry(cb *CircuitBreaker)   { cb.counter.reset() }
func (obs *openBreakerState) onSuccess(cb *CircuitBreaker) {}
func (obs *openBreakerState) onFailure(cb *CircuitBreaker) {}
func (obs *openBreakerState) onExit(cb *CircuitBreaker)    {}
func (obs *openBreakerState) ready() bool                  { return false }

type halfOpenBreakerState struct{}

func (hobs *halfOpenBreakerState) State() breakerStateName      { return breakerHalfOpen }
func (hobs *halfOpenBreakerState) onEntry(cb *CircuitBreaker)   { cb.counter.reset() }
func (hobs *halfOpenBreakerState) onSuccess(cb *CircuitBreaker) { cb.counter.incSuccess() }
func (hobs *halfOpenBreakerState) onFailure(cb *CircuitBreaker) { cb.counter.incFailure() }
func (hobs *halfOpenBreakerState) onExit(cb *CircuitBreaker)    { cb.counter.incFailure() }
func (hobs *halfOpenBreakerState) ready() bool                  { return true }
