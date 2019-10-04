/*
Package retry implements a retry mechanism for test functions.

A simple usage is as simple as

	func TestFooBar(t *testing.T) {
		retry.Run(t, func(t *testing.T) {
			if err := FooBar(); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
*/
package retry

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DefaultPolicy is the default retry policy used with Run.
var DefaultPolicy = &Timer{timeout: 5 * time.Second, sleep: 10 * time.Millisecond}

// Failer represents a partial *testing.SubT.
type Failer interface {
	Log(args ...interface{})
	FailNow()
}

// SubT is a partial implementation of the standard testing testing.SubT.
type SubT struct {
	logs   []string
	failed bool
}

func (t *SubT) reset() {
	t.logs = nil
	t.failed = false
}

func (t *SubT) log(s string) {
	t.logs = append(t.logs, strings.TrimRight(s, "\n"))
}

// Log adds a log line to the current test run.
func (t *SubT) Log(args ...interface{}) {
	t.log(fmt.Sprintln(args...))
}

// Logf adds a formatted log line to the current test run.
func (t *SubT) Logf(format string, args ...interface{}) {
	t.log(fmt.Sprintf(format, args...))
}

// Error adds a log line and fails the current test run.
func (t *SubT) Error(args ...interface{}) {
	t.log(fmt.Sprintln(args...))
	t.Fail()
}

// Errorf adds a formatted log line and fails the current test run.
func (t *SubT) Errorf(format string, args ...interface{}) {
	t.log(fmt.Sprintf(format, args...))
	t.Fail()
}

// Fatal adds a log line, fails the current test run and exits immediately.
func (t *SubT) Fatal(args ...interface{}) {
	t.log(fmt.Sprintln(args...))
	t.FailNow()
}

// Fatalf adds a formatted log line, fails the current test run and exits immediately.
func (t *SubT) Fatalf(format string, args ...interface{}) {
	t.log(fmt.Sprintf(format, args...))
	t.FailNow()
}

// Fail fails the current test run.
func (t *SubT) Fail() {
	t.failed = true
}

// FailNow fails and exits the current test run.
func (t *SubT) FailNow() {
	t.failed = true
	runtime.Goexit()
}

// Run reties fn with the default retry policy.
func Run(t Failer, fn func(t *SubT)) {
	RunWith(t, DefaultPolicy, fn)
}

// RunWith retires fn with policy p.
func RunWith(t Failer, p Policy, fn func(t *SubT)) {
	tt := &SubT{}

	for p.Next() {
		tt.reset()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(tt)
		}()
		wg.Wait()

		if tt.failed {
			continue
		}
		break
	}

	for _, s := range tt.logs {
		t.Log(s)
	}
	if tt.failed {
		t.FailNow()
	}
}

// Policy represents a retry strategy.
type Policy interface {
	// Next determines if the function can be retried. Next is
	// called on the first run, which should be used for any
	// setup that is required.
	Next() bool
}

// Counter is an counter based retry policy.
type Counter struct {
	attempts int
	sleep    time.Duration

	count int
}

// NewCounter returns a counter based retry policy.
func NewCounter(attempts int, sleep time.Duration) *Counter {
	return &Counter{
		attempts: attempts,
		sleep:    sleep,
	}
}

// Next determines if the function can be retried.
func (c *Counter) Next() bool {
	if c.count >= c.attempts {
		return false
	}

	if c.count > 0 {
		time.Sleep(c.sleep)
	}

	c.count++
	return true
}

// Timer is a time based retry policy.
type Timer struct {
	timeout time.Duration
	sleep   time.Duration

	stop time.Time
}

// NewTimer returns a time based retry policy.
func NewTimer(timeout, sleep time.Duration) *Timer {
	return &Timer{
		timeout: timeout,
		sleep:   sleep,
	}
}

// Next determines if the function can be retried.
func (t *Timer) Next() bool {
	if t.stop.IsZero() {
		t.stop = time.Now().Add(t.timeout)
		return true
	}

	if time.Now().After(t.stop) {
		return false
	}

	time.Sleep(t.sleep)
	return true
}
