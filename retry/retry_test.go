package retry_test

import (
	"sync"
	"testing"
	"time"

	"github.com/hamba/testutils/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const timeDeltaAllowed = float64(25 * time.Millisecond)

func TestRun(t *testing.T) {
	mockT := new(MockTestingT)
	mockT.On("Log", []interface{}{"test message"}).Once()
	mockT.On("FailNow").Once()

	var wg sync.WaitGroup

	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		retry.Run(mockT, func(t *retry.SubT) {
			t.Fatal("test message")
		})
	}()
	wg.Wait()
	dur := time.Since(start)

	mockT.AssertExpectations(t)
	assert.InDelta(t, 5*time.Second, dur, timeDeltaAllowed)
}

func TestRunWith_AllowsPassing(t *testing.T) {
	mockT := new(MockTestingT)

	var wg sync.WaitGroup
	var runs int

	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		retry.RunWith(mockT, retry.NewCounter(3, 10*time.Millisecond), func(t *retry.SubT) {
			runs++
		})
	}()
	wg.Wait()
	dur := time.Since(start)

	mockT.AssertExpectations(t)
	assert.Equal(t, 1, runs)
	assert.InDelta(t, 0, dur, timeDeltaAllowed)
}

func TestRunWith_HandlesFailing(t *testing.T) {
	mockT := new(MockTestingT)
	mockT.On("Log", []interface{}{"test message"}).Once()
	mockT.On("FailNow").Once()

	var wg sync.WaitGroup
	var runs int

	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		retry.RunWith(mockT, retry.NewCounter(3, 10*time.Millisecond), func(t *retry.SubT) {
			runs++
			t.Fatal("test message")
		})
	}()
	wg.Wait()
	dur := time.Since(start)

	mockT.AssertExpectations(t)
	assert.Equal(t, 3, runs)
	assert.InDelta(t, 30*time.Millisecond, dur, timeDeltaAllowed)
}

func TestCounter_Next(t *testing.T) {
	p := retry.NewCounter(3, 100*time.Millisecond)

	runs := 0

	start := time.Now()
	for p.Next() {
		runs++
	}
	dur := time.Since(start)

	assert.Equal(t, 3, runs)
	assert.InDelta(t, 200*time.Millisecond, dur, timeDeltaAllowed)
}

func TestTimer_Next(t *testing.T) {
	p := retry.NewTimer(200*time.Millisecond, 100*time.Millisecond)

	runs := 0

	start := time.Now()
	for p.Next() {
		runs++
	}
	dur := time.Since(start)

	assert.Equal(t, 3, runs)
	assert.InDelta(t, 200*time.Millisecond, dur, timeDeltaAllowed)
}

type MockTestingT struct {
	mock.Mock
}

func (m *MockTestingT) Log(args ...interface{}) {
	m.Called(args)
}

func (m *MockTestingT) FailNow() {
	m.Called()
}
