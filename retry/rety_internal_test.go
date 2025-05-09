package retry

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fn         func(t *SubT)
		wantLog    []string
		wantFailed bool
		wantExit   bool
	}{
		{
			name:       "Log",
			fn:         func(t *SubT) { t.Log("test") },
			wantLog:    []string{"test"},
			wantFailed: false,
			wantExit:   false,
		},
		{
			name:       "Logf",
			fn:         func(t *SubT) { t.Logf("%s", "test") },
			wantLog:    []string{"test"},
			wantFailed: false,
			wantExit:   false,
		},
		{
			name:       "Error",
			fn:         func(t *SubT) { t.Error("test") },
			wantLog:    []string{"test"},
			wantFailed: true,
			wantExit:   false,
		},
		{
			name:       "Errorf",
			fn:         func(t *SubT) { t.Errorf("%s", "test") },
			wantLog:    []string{"test"},
			wantFailed: true,
			wantExit:   false,
		},
		{
			name:       "Fatal",
			fn:         func(t *SubT) { t.Fatal("test") },
			wantLog:    []string{"test"},
			wantFailed: true,
			wantExit:   true,
		},
		{
			name:       "Fatalf",
			fn:         func(t *SubT) { t.Fatalf("%s", "test") },
			wantLog:    []string{"test"},
			wantFailed: true,
			wantExit:   true,
		},
		{
			name:       "Fail",
			fn:         func(t *SubT) { t.Fail() },
			wantLog:    []string(nil),
			wantFailed: true,
			wantExit:   false,
		},
		{
			name:       "FailNow",
			fn:         func(t *SubT) { t.FailNow() },
			wantLog:    []string(nil),
			wantFailed: true,
			wantExit:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			exited := true
			retryT := &SubT{}

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				test.fn(retryT)
				exited = false
			}()
			wg.Wait()

			assert.Equal(t, test.wantLog, retryT.logs)
			assert.Equal(t, test.wantFailed, retryT.failed)
			assert.Equal(t, test.wantExit, exited)
		})
	}
}
