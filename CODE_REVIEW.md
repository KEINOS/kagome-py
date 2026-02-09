# Code Review: kagome-py Go Bindings

**Reviewers:** Code quality focus on contribution-friendliness, testability, and maintainability
**Date:** 2026-02-09
**Branch:** initial-implementation
**Coverage:** 70.5% (52+ tests passing)

---

## Executive Summary

The codebase is **well-structured and generally high-quality**, with clear separation of concerns (FFI API, memory management, testing utilities). The test suite is comprehensive and covers edge cases thoroughly.

**Key Strengths:**
- ✅ Clean FFI boundary design with proper memory safety
- ✅ Comprehensive test coverage (52+ tests, G1-G5 + unit tests)
- ✅ Excellent documentation and inline comments
- ✅ Safe nil-pointer handling throughout

**Areas for Enhancement:**
- 🔧 Refactor high-complexity cleanup function
- 🔧 Suppress paralleltest linter for non-parallelizable tests
- 🔧 Improve variable naming in long-scope loops
- 🔧 Allow internal test packages in depguard config
- 🔧 Minor style fixes (return line spacing)

---

## Architecture & Design

### ✅ Strengths

**1. Clear FFI Boundary**
```go
// main.go exports clean C API
//export KagomeInit
//export KagomeDestroy
//export KagomeTokenizeStruct
//export KagomeFreeTokenArray
```
- All public FFI functions are clearly marked with `//export` comments
- Single responsibility: each exported function has one clear purpose
- Proper nil-safety checks at entry points

**2. Memory Safety Pattern**
- ✅ Opaque C handle (malloc'd pointer) prevents Go pointer escape violations
- ✅ Early mutex unlock after map access reduces lock contention
- ✅ Overflow detection before array allocation (prevents DOS)
- ✅ Cleanup on partial allocation failures is comprehensive

**3. Test Architecture**
- ✅ Internal package `cgotest` isolates test helpers from main code
- ✅ No code duplication between `main.go` and test files
- ✅ Comprehensive edge case coverage: empty inputs, large data, unicode, concurrency

### 🔧 Recommendations

**1. Extract Depguard Configuration**
**Issue:** `internal/cgotest` is imported in main_test.go but not in depguard allowlist
**Impact:** Linting error, blocks automation

**Solution:** Update `.golangci.yml` to allow internal test packages:
```yaml
depguard:
  rules:
    main:
      allow:
        - $gostd
        - github.com/ikawaha/kagome/v2
        - github.com/ikawaha/kagome-dict
        - github.com/stretchr/testify
        - github.com/KEINOS/kagome-py/libkagome/internal/cgotest  # Add this
```

**Why:** Makes the internal cgotest package discoverable and allows contributors to run linters without errors.

---

## Code Quality Issues & Fixes

### 1. High Cyclomatic Complexity (cyclop)

**File:** `internal/cgotest/helpers.go:101-166`
**Function:** `CleanupAllocatedTokens()`
**Issue:** Complexity = 15, max = 10

**Current Code:**
```go
func CleanupAllocatedTokens(count int) {
    if count <= 0 { return }

    token := (*C.Token)(C.malloc(...))
    if token == nil { return }
    defer C.free(unsafe.Pointer(token))

    // Allocate 10 strings
    token.surface = C.CString("surface")
    // ... 9 more

    // Loop with conditional frees (5+ conditions)
    for i := 0; i < count; i++ {
        if token.surface != nil { C.free(...) }
        token.surface = C.CString("surface")
    }

    // Final cleanup: 10 separate if statements
    if token.surface != nil { C.free(...) }
    if token.pos1 != nil { C.free(...) }
    // ... 8 more
}
```

**Root Cause:** Repetitive nil-checks for each token field (10 fields × 3 locations = 30+ checks)

**Refactoring Solution:**

Create a helper struct method to reduce complexity:
```go
// tokenStrings is a collection of C strings that need cleanup.
// Reuse the existing tokenStrings type from main.go or create in cgotest.
type tokenStrings struct {
    surface, pos1, pos2, pos3, pos4       *C.char
    conjType, conjForm, baseForm, reading *C.char
    pronunciation                         *C.char
}

// freeAll frees all strings in one call - reduces 10 ifs to 1 call.
func (ts *tokenStrings) freeAll() {
    strs := []*C.char{
        ts.surface, ts.pos1, ts.pos2, ts.pos3, ts.pos4,
        ts.conjType, ts.conjForm, ts.baseForm, ts.reading,
        ts.pronunciation,
    }
    for _, s := range strs {
        if s != nil {
            C.free(unsafe.Pointer(s))
        }
    }
}

// Refactored CleanupAllocatedTokens
func CleanupAllocatedTokens(count int) {
    if count <= 0 {
        return
    }

    token := (*C.Token)(C.malloc(C.size_t(unsafe.Sizeof(C.Token{}))))
    if token == nil {
        return
    }
    defer C.free(unsafe.Pointer(token))

    // Allocate all strings upfront
    ts := tokenStrings{
        surface:       C.CString("surface"),
        pos1:          C.CString("pos1"),
        pos2:          C.CString("pos2"),
        pos3:          C.CString("pos3"),
        pos4:          C.CString("pos4"),
        conjType:      C.CString("conj_type"),
        conjForm:      C.CString("conj_form"),
        baseForm:      C.CString("base_form"),
        reading:       C.CString("reading"),
        pronunciation: C.CString("pronunciation"),
    }
    defer ts.freeAll()

    // Exercise reallocation path
    for range count {
        if ts.surface != nil {
            C.free(unsafe.Pointer(ts.surface))
        }
        ts.surface = C.CString("surface")
    }
}
```

**Benefits:**
- ✅ Reduces cyclomatic complexity from 15 → ~6
- ✅ More testable: can test `freeAll()` separately
- ✅ More maintainable: change to all fields handled in one place
- ✅ DRY: eliminates repeated nil-check patterns

---

### 2. Variable Naming (varnamelen)

**Issue:** Short variable names in long scopes make code harder to scan

**Locations:**
- `main_test.go:486` - `wg` (WaitGroup) used in 15+ lines
- `main_test.go:568` - `wg` (WaitGroup) used in 15+ lines
- `main_test.go:603` - `wg` (WaitGroup) used in 15+ lines
- `main_test.go:658` - `i` loop variable used in 5+ line body
- `helpers.go:127` - `i` loop variable used in 5 line body

**Current:**
```go
var (
    wg           sync.WaitGroup
    successCount atomic.Int32
)

for range numGoroutines {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // ... 10 lines of logic
    }()
}

wg.Wait()
```

**Recommended:**
```go
var (
    goroutineGroup sync.WaitGroup
    successCount   atomic.Int32
)

for range numGoroutines {
    goroutineGroup.Add(1)
    go func() {
        defer goroutineGroup.Done()
        // ... easier to track what's happening
    }()
}

goroutineGroup.Wait()
```

**For loop indices in bodies > 2 lines:**
```go
// Before
for i := 0; i < count; i++ {
    if token.surface != nil { C.free(...) }
    token.surface = C.CString("surface")
}

// After
for iteration := 0; iteration < count; iteration++ {
    if token.surface != nil { C.free(...) }
    token.surface = C.CString("surface")
}
```

---

### 3. Test Parallelization (paralleltest)

**Issue:** Many test functions don't call `t.Parallel()` but linter expects them to

**Context:** Tests modify global `instances` map, so parallelization causes race conditions

**Current State:**
```go
// ✅ Correct - parallel tests that don't touch global state
func TestWouldOverflowTokenAllocation(t *testing.T) {
    t.Parallel()  // Safe
    // ...
}

// ❌ Flagged - but MUST NOT be parallel
func TestKagomeInit(t *testing.T) {
    handle := KagomeInit()  // Modifies global instances map
    // ...
}
```

**Solution:** Suppress warnings with explicit linter directive

**File:** `.golangci.yml`
```yaml
linters-settings:
  paralleltest:
    # Ignore tests that intentionally modify shared state
    # These tests must run serially to prevent race conditions
    ignore-missing:
      - TestKagomeInit
      - TestKagomeDestroy
      - TestG1_.*
      - TestG2_.*
      - TestG3_.*
      - TestG4_.*
      - TestG5_.*
      # Note: Tests that modify global instances map must not be parallelized
```

**OR** Add linter suppression in code:
```go
// TestKagomeInit tests tokenizer initialization.
// Note: Cannot use t.Parallel() - modifies global instances map.
//
//nolint:paralleltest // Intentional: test modifies shared global state
func TestKagomeInit(t *testing.T) {
    handle := KagomeInit()
    // ...
}
```

**Recommendation:** Use the code-level suppression (already mostly done) as it's more explicit about intent.

---

### 4. Return Line Spacing (nlreturn)

**Issue:** Missing blank line before return statement in one-function chains

**File:** `internal/cgotest/helpers.go`

**Current:**
```go
// Line 48-50 (helpers.go)
func TokenizeString(handle unsafe.Pointer, text string) int {
    cStr := C.CString(text)
    defer C.free(unsafe.Pointer(cStr))  // <-- no blank line
    return 0  // or processing continues
}

// Line 72-74
func HandleExists(handle unsafe.Pointer) bool {
    defer C.free(unsafe.Pointer(cStr))
    arr := C.KagomeTokenizeStruct(handle, cStr)  // <-- no blank line
    if arr == nil {
        return false
    }
}
```

**Fix:** Add blank line before return
```go
func TokenizeString(handle unsafe.Pointer, text string) int {
    cStr := C.CString(text)
    defer C.free(unsafe.Pointer(cStr))

    arr := C.KagomeTokenizeStruct(handle, cStr)
    if arr == nil {
        return 0
    }

    count := int(arr.length)
    C.KagomeFreeTokenArray(arr)

    return count
}
```

---

### 5. Modern Range Syntax (intrange)

**Go 1.22+** supports simplified range syntax:

**File:** `internal/cgotest/helpers.go:127`

**Current:**
```go
for i := 0; i < count; i++ {
    if token.surface != nil {
        C.free(unsafe.Pointer(token.surface))
    }
    token.surface = C.CString("surface")
}
```

**Modern (Go 1.22+):**
```go
for range count {
    if token.surface != nil {
        C.free(unsafe.Pointer(token.surface))
    }
    token.surface = C.CString("surface")
}
```

**Note:** Only applies when loop variable `i` is unused.

---

## Contribution Friendliness Assessment

### ✅ Excellent
- **Clear FFI boundary:** Exported functions are obviously C-facing
- **Comprehensive docs:** Each function has purpose statement and safety notes
- **Inline comments:** Complex operations (overflow checks, early unlock) are explained
- **Test patterns:** Easy to add more G1-G5 tests following existing patterns

### 🔧 Can Improve
- **Setup friction:** Linter errors block first-time contributor runs
  - **Fix:** Add internal package to depguard config (5 min fix)
- **Complexity barrier:** `CleanupAllocatedTokens` is hard to modify
  - **Fix:** Extract cleanup helper (30 min refactor, highly testable)
- **Variable names:** Some names require context to understand
  - **Fix:** Rename `wg` → `goroutineGroup` (15 min, pure clarity win)

---

## Enhancement Opportunities

### Easy Wins (contrib-ready)

1. **Add go.mod documentation**
   - Current: Minimal `module` and `go` directives
   - Suggested: Add `// Comment` with package purpose

2. **Document test helper package**
   - Create `internal/cgotest/README.md`
   - Explain why it's internal (isolation) and how to add new helpers

3. **Add CONTRIBUTING.md section**
   - How to run tests locally: `go test -cover ./...`
   - How to run linter: `golangci-lint run --fix`
   - Known linting suppressions and why

### Medium Refactors (testability improvement)

1. **Restructure CleanupAllocatedTokens** (see section above)
   - Makes function easier to understand
   - Enables targeted unit testing of cleanup logic

2. **Extract token string marshaling**
   - Currently spread across `allocateTokenStrings()` and `checkAndStoreToken()`
   - Create explicit conversion function: `marshallToken(*tokenizer.Token) (C.Token, error)`
   - More testable, easier to modify for future token fields

### Larger Enhancements (future versions)

1. **Parallel-safe test refactor**
   - Use sync.Once or test fixtures to share instances between parallel subtests
   - Would enable `t.Parallel()` on all tests

2. **Benchmarks**
   - Add `BenchmarkKagomeTokenize()` to track performance
   - Compare against pure Go Kagome

3. **Example code**
   - Add `example_test.go` showing FFI usage for new contributors
   - Tests must compile and run to stay in sync

---

## Testing Assessment

### Coverage Analysis
- **Statement coverage:** 70.5% (good for production code)
- **Missing coverage:** Mostly error paths and edge allocation failures (acceptable)

### Test Quality

**Strengths:**
- ✅ **Edge cases covered:** Empty, very long, unicode, boundary conditions
- ✅ **Concurrency tested:** 1000+ goroutine tests, handle persistence
- ✅ **Error recovery:** Tests after failures, partial allocations
- ✅ **Modern patterns:** Uses testify/require, t.Cleanup(), atomic operations

**Improvement Opportunities:**
- 🔧 Add benchmarks for performance tracking
- 🔧 Add table-driven tests for variations (already partially done well)
- 🔧 Document test strategy in README

---

## Security Considerations

### Memory Safety
- ✅ All C strings have explicit freeing
- ✅ Overflow checks before allocation
- ✅ Nil-pointer handling throughout
- ✅ No buffer overruns possible (uses C.malloc/free)

### FFI Boundary
- ✅ No Go pointers passed to C (opaque handles only)
- ✅ Input validation on all exported functions
- ✅ String conversions use C.GoString/C.CString (safe)

### Concurrency
- ✅ Mutex protects instances map
- ✅ Early unlock reduces contention
- ✅ Atomic operations for counters
- ✅ No shared mutable state in token arrays

---

## Checklist for Contributor Prep

- [ ] Update `.golangci.yml` to allow internal/cgotest in depguard (5 min)
- [ ] Refactor `CleanupAllocatedTokens` with helper struct (30 min)
- [ ] Rename `wg` → `goroutineGroup`, `i` → `iteration` in test loops (15 min)
- [ ] Add blank lines before returns in helpers.go (5 min)
- [ ] Add linter suppression to paralleltest tests (5 min)
- [ ] Create `internal/cgotest/README.md` (15 min)
- [ ] Add CONTRIBUTING.md section for Go development (20 min)
- [ ] Run full test suite: `go test -cover ./...` (2 min)
- [ ] Run linter: `golangci-lint run` (should pass with fixes above) (2 min)

**Total preparation time:** ~90 min
**Difficulty:** Low-Medium (mostly code clarity improvements)

---

## Conclusion

The codebase is **production-ready with excellent foundations**. The main improvements are around contributor experience (linter config) and code clarity (variable names, complexity), not correctness or safety.

**For next release/v2.1+:**
1. Address linter config to unblock automation
2. Refactor `CleanupAllocatedTokens` for maintainability
3. Add contributor documentation
4. Consider parallel test refactor (nice-to-have, lower priority)

**Overall Quality Score: 8.5/10** ⭐
- Functionality: 9.5/10
- Testing: 9/10
- Maintainability: 7.5/10 (complexity issue)
- Contribution-friendliness: 8/10 (linter config gap)
