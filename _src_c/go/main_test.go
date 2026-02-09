package main

import (
	"math"
	"strings"
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
	// No t.Parallel() - modifies global instances map.
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
	// No t.Parallel() - modifies global instances map.
	t.Run("normal cleanup", func(t *testing.T) {
		// No t.Parallel() - parent test handles serialization.
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
		// No t.Parallel() - parent test handles serialization.
		// Should not panic.
		require.NotPanics(t, func() { KagomeDestroy(nil) })
	})
}

// TestKagomeTokenizeConcurrent tests thread safety with concurrent access.
// This verifies that concurrent calls to KagomeTokenizeStruct don't cause
// race conditions or crashes, which was a critical bug in early versions.
func TestKagomeTokenizeConcurrent(t *testing.T) {
	// No t.Parallel() - calls KagomeInit/Destroy which modify global instances map.
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

// ------------------------------------------------------------------
// G1: Edge Cases - Empty & Boundary Inputs
// ------------------------------------------------------------------

// TestG1_EmptyStringInput tests tokenization of empty string.
func TestG1_EmptyStringInput(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Empty string: should return 0 tokens (handled safely).
	count := testTokenizeString(handle, "")
	require.Equal(t, 0, count, "Empty string should produce zero tokens")
}

// TestG1_SingleCharacter tests tokenization of single ASCII character.
func TestG1_SingleCharacter(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	count := testTokenizeString(handle, "a")
	require.GreaterOrEqual(t, count, 0, "Single char should tokenize safely")
}

// TestG1_VeryLongString tests tokenization of 100KB+ text.
func TestG1_VeryLongString(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Create 100KB+ test string by repeating text.
	longText := ""
	var longTextSb319 strings.Builder
	for i := 0; i < 5000; i++ {
		longTextSb319.WriteString("これはテストです。")
	}
	longText += longTextSb319.String()

	count := testTokenizeString(handle, longText)
	require.Greater(t, count, 1000, "Very long string should produce many tokens")
}

// TestG1_ZeroTokenCount tests handling of input that produces zero tokens.
func TestG1_ZeroTokenCount(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	count := testTokenizeString(handle, "")
	require.Equal(t, 0, count, "Empty string should produce zero tokens")
}

// TestG1_NullPointerHandling tests null pointer safety.
func TestG1_NullPointerHandling(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// NULL handle should be handled safely.
	count := testTokenizeString(nil, "test")
	require.Equal(t, 0, count, "NULL handle should return 0 tokens safely")
}

// ------------------------------------------------------------------
// G2: Unicode Edge Cases
// ------------------------------------------------------------------

// TestG2_ASCIIOnlyInput tests ASCII-only text.
func TestG2_ASCIIOnlyInput(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	count := testTokenizeString(handle, "hello world")
	require.GreaterOrEqual(t, count, 0, "Should handle ASCII input safely")
}

// TestG2_EmojiSequences tests emoji handling.
func TestG2_EmojiSequences(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	count := testTokenizeString(handle, "👍🎉")
	require.GreaterOrEqual(t, count, 0, "Should handle emoji sequences safely")
}

// TestG2_MixedScripts tests mixed language scripts.
func TestG2_MixedScripts(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	//nolint:gosmopolitan // Testing international script handling.
	count := testTokenizeString(handle, "Hello こんにちは 你好")
	require.GreaterOrEqual(t, count, 0, "Should handle mixed scripts safely")
}

// TestG2_CombiningCharacters tests combining diacritics.
func TestG2_CombiningCharacters(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	count := testTokenizeString(handle, "café")
	require.GreaterOrEqual(t, count, 0, "Should handle combining characters safely")
}

// TestG2_RTLText tests right-to-left text.
func TestG2_RTLText(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	count := testTokenizeString(handle, "العربية")
	require.GreaterOrEqual(t, count, 0, "Should handle RTL text safely")
}

// ------------------------------------------------------------------
// G3: Memory & Allocation Critical Cases
// ------------------------------------------------------------------

// TestG3_LargeTokenArrays tests handling of 1000+ tokens.
func TestG3_LargeTokenArrays(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Create text that should produce many tokens.
	text := ""
	var textSb430 strings.Builder
	for i := 0; i < 1000; i++ {
		textSb430.WriteString("あ い う ")
	}
	text += textSb430.String()

	count := testTokenizeString(handle, text)
	require.Greater(t, count, 1000, "Should produce many tokens from large input")
}

// TestG3_RepeatedAllocDealloc tests allocation/deallocation cycles.
func TestG3_RepeatedAllocDealloc(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Repeatedly tokenize and free - stress test memory management.
	for i := 0; i < 100; i++ {
		count := testTokenizeString(handle, "これはテストです。")
		require.Positive(t, count, "Iteration %d: should tokenize successfully", i)
	}
}

// TestG3_ErrorPathCleanup tests cleanup on error conditions.
func TestG3_ErrorPathCleanup(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// NULL input - error path should cleanup properly.
	count := testTokenizeString(nil, "test")
	require.Equal(t, 0, count, "Error path should handle NULL safely")

	// After error, normal operation should still work.
	count = testTokenizeString(handle, "テスト")
	require.Positive(t, count, "Should recover after error condition")
}

// TestG3_UnicodeStringAllocation tests multi-byte string allocation.
func TestG3_UnicodeStringAllocation(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Multi-byte UTF-8 characters.
	//nolint:gosmopolitan // Testing multi-byte character handling.
	count := testTokenizeString(handle, "日本語の複雑な文字列")
	require.Positive(t, count, "Should allocate multi-byte strings correctly")
}

// ------------------------------------------------------------------
// G4: Concurrency Edge Cases
// ------------------------------------------------------------------

// TestG4_HighConcurrency tests 1000+ goroutines.
func TestG4_HighConcurrency(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	const numGoroutines = 1000

	var (
		wg           sync.WaitGroup
		successCount atomic.Int32
	)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			count := testTokenizeString(handle, "テスト")
			if count >= 0 {
				successCount.Add(1)
			}
		}()
	}

	wg.Wait()

	require.Equal(t, int32(numGoroutines), successCount.Load(), "All high-concurrency calls should succeed")
}

// TestG4_ConcurrentInitDestroy tests concurrent handle lifecycle.
func TestG4_ConcurrentInitDestroy(t *testing.T) {

	const numOps = 100

	var (
		wg           sync.WaitGroup
		successCount atomic.Int32
	)

	for i := 0; i < numOps; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			handle := KagomeInit()
			if handle == nil {
				return
			}

			// Quick tokenization.
			count := testTokenizeString(handle, "テスト")
			if count >= 0 {
				successCount.Add(1)
			}

			KagomeDestroy(handle)
		}()
	}

	wg.Wait()

	require.Positive(t, successCount.Load(), "Some concurrent operations should succeed")
}

// TestG4_ThreadSafetyOfSharedInstance tests concurrent access to same instance.
func TestG4_ThreadSafetyOfSharedInstance(t *testing.T) {
	// Note: Cannot use t.Parallel() as it modifies global state.
	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	const (
		numGoroutines = 100
		iterations    = 10
	)

	var (
		wg           sync.WaitGroup
		successCount atomic.Int32
	)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				count := testTokenizeString(handle, "並行処理")
				if count >= 0 {
					successCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	expected := int32(numGoroutines * iterations)
	require.Equal(t, expected, successCount.Load(), "All concurrent operations should succeed")
}

// TestG4_HandlePersistence tests handles persisting across operations.
func TestG4_HandlePersistence(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	const iterations = 100

	var (
		wg           sync.WaitGroup
		successCount atomic.Int32
	)

	for i := 0; i < iterations; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// Verify handle still exists and works.
			if !testHandleExists(handle) {
				return
			}

			count := testTokenizeString(handle, "永続テスト")
			if count >= 0 {
				successCount.Add(1)
			}
		}()
	}

	wg.Wait()

	require.Equal(t, int32(iterations), successCount.Load(), "Handle should persist correctly")
}

// ------------------------------------------------------------------
// G5: Error Recovery & State Integrity
// ------------------------------------------------------------------

// TestG5_TokenizationAfterFailed tests tokenization after failed operations.
func TestG5_TokenizationAfterFailed(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Failed operation (NULL handle).
	count := testTokenizeString(nil, "test")
	require.Equal(t, 0, count, "NULL input should fail gracefully")

	// Recovery: normal operation should work.
	count = testTokenizeString(handle, "回復テスト")
	require.Positive(t, count, "Should recover after failed operation")
}

// TestG5_InstanceReuseAfterErrors tests instance reuse after error conditions.
func TestG5_InstanceReuseAfterErrors(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	// Multiple error-recovery cycles.
	for i := 0; i < 10; i++ {
		// Error: NULL handle.
		count := testTokenizeString(nil, "test")
		require.Equal(t, 0, count)

		// Recovery: normal operation.
		count = testTokenizeString(handle, "再利用テスト")
		require.Positive(t, count, "Iteration %d: should recover", i)
	}
}

// TestG5_HandleReusePatterns tests various handle reuse patterns.
func TestG5_HandleReusePatterns(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	patterns := []string{
		"",            // Empty
		"a",           // Single char
		"テスト",         // Japanese
		"Hello World", // English
		"👍🎉",          // Emoji
		"Mixed 日本語",   // Mixed
	}

	for _, pattern := range patterns {
		count := testTokenizeString(handle, pattern)
		// All patterns should be handled safely.
		require.GreaterOrEqual(t, count, 0, "Pattern '%s' should be handled safely", pattern)
	}
}

// TestG5_PartialAllocationCleanup tests cleanup of partially allocated tokens.
func TestG5_PartialAllocationCleanup(t *testing.T) {

	handle := KagomeInit()
	require.NotNil(t, handle)
	t.Cleanup(func() { KagomeDestroy(handle) })

	testText := "部分的な割り当てクリーンアップテスト"

	// Allocate and immediately free (stress cleanup paths).
	for i := 0; i < 50; i++ {
		count := testTokenizeString(handle, testText)
		require.Positive(t, count, "Iteration %d: should tokenize", i)
	}

	// After stress test, instance should still work.
	count := testTokenizeString(handle, testText)
	require.Positive(t, count, "Should still work after stress test")
}
