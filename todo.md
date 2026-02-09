# TODO & Issues

Issues, improvements, and testing tasks for kagome-py.

## Status Summary

### v2.10.3 Release - READY [COMPLETE]

All MUST FIX items done. Release is unblocked.

### Testing & Robustness - v2.1+ [COMPLETE ✅]

All comprehensive edge case tests implemented (55 total tests, 100% pass rate).
- ✅ G1-G5 Go tests: 24 tests
- ✅ P1-P6 Python tests: 31 tests

### Go Code Quality & Maintainability - v2.1+ [IN PROGRESS]

Code review identified linting issues and refactoring opportunities for contributor-friendliness.
**Goal:** Make codebase easy to enhance and extend.

---

## v2.10.3 Release Status

### MUST FIX BEFORE RELEASE - ALL COMPLETE!

1. ~~Fix `macos-12` → `macos-15-intel` in publish.yml (#1)~~ [DONE - c9ea815]
2. ~~Single-source version (remove duplication #2)~~ [DONE - 1ddddf4]
3. ~~Add `__eq__` / `__repr__` to Token (#3)~~ [DONE - 5b3e347]
4. ~~Add Go tests to publish.yml (#4)~~ [DONE - 1ef083f]
5. ~~Fix Windows Arm64 claim in README (#18)~~ [DONE - f338df5]

---

## Testing & Robustness for v2.1+ [COMPLETE ✅]

Comprehensive test coverage organized by language: Go first, then Python.

### GO TESTS: G1-G5 [COMPLETE ✅]

**File:** `_src_c/go/main_test.go` - 24 comprehensive tests

- ✅ **G1** (5 tests): Edge cases - empty, single char, 100KB+ text, boundaries, null pointers
- ✅ **G2** (5 tests): Unicode - ASCII, emoji, mixed scripts, combining chars, RTL text
- ✅ **G3** (4 tests): Memory - large arrays (1000+), allocation cycles, error cleanup, unicode
- ✅ **G4** (5 tests): Concurrency - 1000+ goroutines, concurrent init/destroy, shared instance, persistence
- ✅ **G5** (5 tests): Error recovery - after failures, instance reuse, handle patterns, partial cleanup

**Status:** All 24 tests passing ✅ | Coverage: 70.5% statements | Commit: 8d0e7d4

### PYTHON TESTS: P1-P6 [COMPLETE ✅]

**File:** `tests/libkagome_test.py` - 31 comprehensive tests

- ✅ **P1** (5 tests): Edge cases - empty, single char, 100KB+ text, whitespace
- ✅ **P2** (5 tests): Unicode - ASCII, emoji, mixed scripts, diacritics, RTL
- ✅ **P3** (6 tests): wakati() method - empty, single word, sentence, ASCII, mixed, consistency
- ✅ **P4** (6 tests): Token equality - __eq__ with None/str/Token, repr, list membership
- ✅ **P5** (3 tests): Multiple instances - independent, reuse, memory
- ✅ **P6** (4 tests): Version & metadata - __version__, __all__, docstring, public API

**Status:** All 31 tests passing ✅ | 100% success rate | Commit: ee592a9

---

## Go Code Quality & Maintainability (v2.1+)

### Linting & Code Clarity [IN PROGRESS]

**Goal:** Fix golangci-lint issues and improve contributor-friendliness

#### PRIORITY 1: Fix Linting Errors (automation blocker)

**File:** `_src_c/go/.golangci.yml`

- [ ] Add `github.com/KEINOS/kagome-py/libkagome/internal/cgotest` to depguard allowlist
  - **Impact:** Unblocks CI/CD automation
  - **Effort:** 5 minutes
  - **Status:** Ready to fix

**Issues Addressed:** 1 depguard error

#### PRIORITY 2: Reduce Cyclomatic Complexity

**File:** `_src_c/go/internal/cgotest/helpers.go:101`

- [ ] Refactor `CleanupAllocatedTokens()` - complexity 15 → target 6
  - Extract `tokenStrings` struct with `freeAll()` method
  - Reduces 10 separate nil-checks to 1 call
  - Improves testability and maintainability
  - **Effort:** 30 minutes
  - **Impact:** Makes function easier to modify

**Issues Addressed:** 1 cyclop error

#### PRIORITY 3: Improve Variable Naming (readability)

**File:** `_src_c/go/main_test.go` + `helpers.go`

- [ ] Rename `wg` → `goroutineGroup` (3 locations, 15+ line scopes)
  - **Effort:** 15 minutes
  - **Impact:** Code is easier to scan and understand

- [ ] Rename loop variable `i` → `iteration` (where body > 2 lines)
  - **Effort:** 10 minutes
  - **Impact:** Clearer intent in complex loops

**Issues Addressed:** 4 varnamelen warnings

#### PRIORITY 4: Code Style Consistency

**File:** `_src_c/go/internal/cgotest/helpers.go`

- [ ] Add blank lines before return statements (nlreturn)
  - Locations: lines 49-50, 72-74
  - **Effort:** 5 minutes

- [ ] Use Go 1.22+ range syntax where loop var unused
  - Location: line 127 `for i := 0; i < count; i++` → `for range count`
  - **Effort:** 5 minutes

**Issues Addressed:** 4 style warnings (nlreturn + intrange)

#### PRIORITY 5: Suppress Intentional Paralleltest Warnings

**File:** `.golangci.yml` or individual test functions

- [ ] Add linter suppression for tests modifying global instances map
  - **Why:** Tests MUST NOT use t.Parallel() due to shared state
  - **Approach:** Use `//nolint:paralleltest` on test functions
  - **Effort:** 10 minutes
  - **Impact:** Silences expected warnings, documents intent

**Issues Addressed:** 27 paralleltest warnings

#### PRIORITY 6: Minor Style Fixes

**File:** Various

- [ ] Add gosmopolitan nolint comments (already mostly done)
  - Already suppressed in G2/G5 tests
  - **Effort:** 0 minutes (done)

**Total Effort:** ~90 minutes to complete all issues
**Expected Result:** `golangci-lint run` with 0 issues

### Code Review Documents

- [ ] Commit CODE_REVIEW.md with detailed analysis
  - Includes architecture assessment, security audit
  - Contribution readiness checklist
  - Enhancement opportunities for v2.2+
  - **Status:** Ready to commit

---

## Deferred Items (v2.2+)

### Go Code Quality (from v2.1)

- Refactor `CleanupAllocatedTokens()` complexity (deferred from v2.1)
- Add internal test package documentation
- Parallel test refactor (advanced: enable t.Parallel() on all tests)

### CI/CD & Refactoring

- Refactor CI/CD: Extract reusable build workflows
- Python version matrix: Test 3.10/3.11/3.12/3.13
- Migrate test framework: Hand-rolled → pytest
- macOS deployment target verification
- Add `-pthread` link flag on Linux
- Thread-safety documentation

### Code Enhancements

- Context manager protocol (`with Kagome() as k:`)
- Thread safety enforcement (locking if needed)
- CLI entry point (`python -m libkagome`)
- Type hints with `py.typed` marker
- Explicit `__all__` exports
- Add example_test.go for contributor reference

---

## Implementation Notes

**Testing Strategy (v2.1 - COMPLETE):**

1. Go tests establish correctness and safety at FFI boundary
2. Python tests verify language binding behavior
3. Edge cases discovered in one layer inform the other
4. Priority order: CRITICAL → HIGH → MEDIUM → LOW

**Test Organization:**

- Go: Organized in `_src_c/go/main_test.go` by G1-G5 categories
  - Uses testify/require for assertions
  - Comprehensive helper functions in `internal/cgotest/helpers.go`
  - 24 tests covering all critical paths

- Python: Organized in `tests/libkagome_test.py` by P1-P6 categories
  - Hand-rolled test framework (pytest migration in v2.2+)
  - 31 tests covering Python-specific behavior
  - Original integration test preserved

**Running Tests:**

```bash
# Go with coverage
cd _src_c/go
go test -cover ./...          # Run with coverage report
golangci-lint run             # Check code quality

# Python (existing)
PYTHONPATH=./src python3 tests/libkagome_test.py

# Code review
cat CODE_REVIEW.md            # Detailed analysis and recommendations
```

**Code Quality Standards (v2.1):**

- golangci-lint: strict checking, all issues must be fixed
- Test coverage: minimum 70% (current: 70.5%)
- Memory safety: no resource leaks allowed
- Security: FFI boundary carefully audited
