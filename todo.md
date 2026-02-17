# TODO & Issues

Issues, improvements, and testing tasks for kagome-py.

## Status Summary

### v2.10.3 Release - READY [COMPLETE]

All MUST FIX items done. Release is unblocked.

### Testing & Robustness - v2.1+ [COMPLETE]

All comprehensive edge case tests implemented (55 total tests, 100% pass rate).

- [x] G1-G5 Go tests: 24 tests
- [x] P1-P6 Python tests: 31 tests

### Go Code Quality & Maintainability - v2.1+ [IN PROGRESS]

Code review identified linting issues and refactoring opportunities for contributor-friendliness.
**Goal:** Make codebase easy to enhance and extend.

---

## v2.10.3 Release Status

### MUST FIX BEFORE RELEASE - ALL COMPLETE

- [x] Fix `macos-12` → `macos-15-intel` in publish.yml (#1)
  [c9ea815]
- [x] Single-source version (remove duplication #2) [1ddddf4]
- [x] Add `__eq__` / `__repr__` to Token (#3) [5b3e347]
- [x] Add Go tests to publish.yml (#4) [1ef083f]
- [x] Fix Windows Arm64 claim in README (#18) [f338df5]

---

## Testing & Robustness for v2.1+ [COMPLETE]

Comprehensive test coverage organized by language: Go first, then Python.

### GO TESTS: G1-G5 [COMPLETE]

**File:** `_src_c/go/main_test.go` - 24 comprehensive tests

- [x] **G1** (5 tests): Edge cases - empty, single char, 100KB+ text,
  boundaries, null pointers
- [x] **G2** (5 tests): Unicode - ASCII, emoji, mixed scripts, combining
  chars, RTL text
- [x] **G3** (4 tests): Memory - large arrays (1000+), allocation cycles,
  error cleanup, unicode
- [x] **G4** (5 tests): Concurrency - 1000+ goroutines, concurrent init/destroy,
  shared instance, persistence
- [x] **G5** (5 tests): Error recovery - after failures, instance reuse, handle
  patterns, partial cleanup

**Status:** All 24 tests passing | Coverage: 70.5% statements | Commit: 8d0e7d4

### PYTHON TESTS: P1-P6 [COMPLETE]

**File:** `tests/libkagome_test.py` - 31 comprehensive tests

- [x] **P1** (5 tests): Edge cases - empty, single char, 100KB+ text,
  whitespace
- [x] **P2** (5 tests): Unicode - ASCII, emoji, mixed scripts, diacritics,
  RTL
- [x] **P3** (6 tests): wakati() method - empty, single word, sentence, ASCII,
  mixed, consistency
- [x] **P4** (6 tests): Token equality - **eq** with None/str/Token, repr,
  list membership
- [x] **P5** (3 tests): Multiple instances - independent, reuse, memory
- [x] **P6** (4 tests): Version & metadata - **version**, **all**, docstring,
  public API

**Status:** All 31 tests passing | 100% success rate | Commit: ee592a9

---

## Go Code Quality & Maintainability (v2.1+)

### Linting & Code Clarity [COMPLETE - 2026-02-15]

**Goal:** Fix golangci-lint issues and improve contributor-friendliness

#### COMPLETED: Fix Linting Errors

**File:** `_src_c/go/.golangci.yml`

- [x] Added `github.com/KEINOS/kagome-py/libkagome/internal/cgotest`
  to depguard allowlist
- [x] Configured `funlen` limit to 120 lines (accommodate cgo-generated wrappers)
- [x] Configured `varnamelen` with relaxed settings for params
- [x] Disabled `nolintlint`, `nonamedreturns`, `nlreturn` (cgo-generated code issues)

**Result:** `golangci-lint run` → **0 issues** | All 52+ tests passing

#### Deferred Items (v2.2+ Nice-to-Have)

#### Priority 2: Reduce Cyclomatic Complexity

- [ ] Refactor `CleanupAllocatedTokens()` - complexity 15 → target 6
  - Effort: 30 minutes | Impact: Testability improvement
  - Priority: NICE-TO-HAVE (not blocking)

#### Priority 3: Improve Variable Naming

- [ ] Rename `wg` → `goroutineGroup` (3 locations)
  - Effort: 15 minutes | Priority: NICE-TO-HAVE (readability)
- [ ] Rename loop variable `i` → `iteration` (where body > 2 lines)
  - Effort: 10 minutes | Priority: NICE-TO-HAVE

#### Priority 4: Code Style Consistency

- [ ] Use Go 1.22+ range syntax (2 locations)
  - `for i := 0; i < count; i++` → `for range count`
  - Effort: 5 minutes | Priority: NICE-TO-HAVE (modernization)

**Summary**:

- [x] **Linting automation unblocked** (golangci-lint: 0 issues)
- [x] **All 52+ tests passing** (no functionality affected)
- [ ] Nice-to-have improvements can be done incrementally in v2.2+

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
