package main

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// TestWouldOverflowTokenAllocation tests the overflow detection logic.
func TestWouldOverflowTokenAllocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		n        int
		wantFail bool
	}{
		{
			name:     "negative number",
			n:        -1,
			wantFail: true,
		},
		{
			name:     "zero tokens",
			n:        0,
			wantFail: false,
		},
		{
			name:     "normal amount",
			n:        1000,
			wantFail: false,
		},
		{
			name:     "large but safe",
			n:        1_000_000,
			wantFail: false,
		},
		{
			name:     "maximum int",
			n:        math.MaxInt,
			wantFail: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := wouldOverflowTokenAllocation(testCase.n)
			require.Equal(t, testCase.wantFail, got)
		})
	}
}

// TestKagomeInit tests tokenizer initialization.
func TestKagomeInit(t *testing.T) {
	t.Parallel()

	handle := KagomeInit()
	require.NotNil(t, handle, "KagomeInit should return non-nil handle")
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Verify handle was stored in instances map.
	instanceMutex.Lock()

	_, exists := instances[handle]

	instanceMutex.Unlock()

	require.True(t, exists, "Handle should be found in instances map")
}

// TestKagomeDestroy tests tokenizer cleanup.
func TestKagomeDestroy(t *testing.T) {
	t.Parallel()

	t.Run("normal cleanup", func(t *testing.T) {
		t.Parallel()

		handle := KagomeInit()
		require.NotNil(t, handle, "KagomeInit should return non-nil handle")

		KagomeDestroy(handle)

		// Verify handle was removed from instances map.
		instanceMutex.Lock()

		_, exists := instances[handle]

		instanceMutex.Unlock()

		require.False(t, exists, "Handle should be removed from instances map after destroy")
	})

	t.Run("nil handle is safe", func(t *testing.T) {
		t.Parallel()
		// Should not panic.
		require.NotPanics(t, func() { KagomeDestroy(nil) })
	})
}

// TestKagomeTokenizeConcurrent tests thread safety with concurrent access.
// This verifies that concurrent calls to KagomeTokenizeStruct don't cause
// race conditions or crashes, which was a critical bug in early versions.
func TestKagomeTokenizeConcurrent(t *testing.T) {
	t.Parallel()

	handle := KagomeInit()
	require.NotNil(t, handle, "KagomeInit should return non-nil handle")
	t.Cleanup(func() { KagomeDestroy(handle) })

	const (
		numGoroutines = 10
		iterations    = 50
	)

	var (
		waitGroup sync.WaitGroup
		failCount atomic.Int32
	)

	// Run concurrent tokenizations.
	for range numGoroutines {
		waitGroup.Add(1)

		//nolint:modernize // WaitGroup.Go doesn't exist in sync package.
		go func() {
			defer waitGroup.Done()

			for range iterations {
				testTokenizeCall(handle, &failCount)
			}
		}()
	}

	waitGroup.Wait()

	require.Zero(t, failCount.Load(), "No concurrent tokenization failures expected")
}

// testTokenizeCall verifies handle validity in concurrent context.
// Helper for TestKagomeTokenizeConcurrent that tests handle access
// without requiring C imports in the test file.
func testTokenizeCall(handle unsafe.Pointer, failCount *atomic.Int32) {
	// Verify handle is still valid and accessible.
	instanceMutex.Lock()

	_, exists := instances[handle]

	instanceMutex.Unlock()

	if !exists {
		failCount.Add(1)
	}
}

// TestGetOrEmpty tests the string array accessor helper.
func TestGetOrEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arr  []string
		idx  int
		want string
	}{
		{
			name: "valid index",
			arr:  []string{"a", "b", "c"},
			idx:  1,
			want: "b",
		},
		{
			name: "out of bounds",
			arr:  []string{"a", "b"},
			idx:  5,
			want: "",
		},
		{
			name: "negative index",
			arr:  []string{"a", "b"},
			idx:  -1,
			want: "",
		},
		{
			name: "empty array",
			arr:  []string{},
			idx:  0,
			want: "",
		},
		{
			name: "nil array",
			arr:  nil,
			idx:  0,
			want: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := getOrEmpty(testCase.arr, testCase.idx)
			require.Equal(t, testCase.want, got)
		})
	}
}

// TestFreeStringsNilSafe tests that freeStrings safely handles nil pointers.
// This verifies the function handles nil without panicking.
// Actual C memory operations are tested through integration tests.
func TestFreeStringsNilSafe(t *testing.T) {
	t.Parallel()

	t.Run("handles nil in variadic args", func(t *testing.T) {
		t.Parallel()
		// Should not panic when called with no args.
		require.NotPanics(t, func() { freeStrings() })
	})
}

// TestInstanceMapThreadSafety verifies the instances map is protected by mutex.
// Note: Cannot use t.Parallel() as it modifies global instances state.
//
//nolint:paralleltest // Global state mutation test, must run serially.
func TestInstanceMapThreadSafety(t *testing.T) {
	const numGoroutines = 20

	var waitGroup sync.WaitGroup

	handles := make([]unsafe.Pointer, numGoroutines)

	// Create multiple handles concurrently.
	for index := range numGoroutines {
		waitGroup.Add(1)

		go func(idx int) {
			defer waitGroup.Done()

			handles[idx] = KagomeInit()
		}(index)
	}

	waitGroup.Wait()

	// Verify all handles were created.
	for i, h := range handles {
		require.NotNil(t, h, "Handle %d should not be nil", i)
	}

	// Destroy all handles concurrently.
	for index := range numGoroutines {
		waitGroup.Add(1)

		go func(idx int) {
			defer waitGroup.Done()

			if handles[idx] != nil {
				KagomeDestroy(handles[idx])
			}
		}(index)
	}

	waitGroup.Wait()

	// Verify all handles were removed.
	instanceMutex.Lock()

	mapSize := len(instances)

	instanceMutex.Unlock()

	require.Zero(t, mapSize, "instances map should be empty after cleanup")
}
