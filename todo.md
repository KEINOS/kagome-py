# TODO & Issues

Issues, improvements, and testing tasks for kagome-py.

## Status Summary

### v2.10.3 Release - READY [COMPLETE]
All MUST FIX items done. Release is unblocked.

### Testing & Robustness - v2.1+ [NEW]
Comprehensive edge case, niche, and critical tests organized by language.

---

## v2.10.3 Release Status

### MUST FIX BEFORE RELEASE - ALL COMPLETE!

1. ~~Fix `macos-12` → `macos-15-intel` in publish.yml (#1)~~ [DONE - c9ea815]
2. ~~Single-source version (remove duplication #2)~~ [DONE - 1ddddf4]
3. ~~Add `__eq__` / `__repr__` to Token (#3)~~ [DONE - 5b3e347]
4. ~~Add Go tests to publish.yml (#4)~~ [DONE - 1ef083f]
5. ~~Fix Windows Arm64 claim in README (#18)~~ [DONE - f338df5]

---

## Testing & Robustness for v2.1+ (NEW)

Comprehensive test coverage organized by language: Go first, then Python.

### GO TESTS: Edge Cases & Robustness

#### G1. Go: Edge Cases - Empty & Boundary Inputs
**File:** `_src_c/go/main_test.go`

- Test empty string input: `KagomeTokenizeStruct(handle, "")`
- Test single character: `KagomeTokenizeStruct(handle, "a")`
- Test very long string (100KB+)
- Test max int boundaries for token counts
- Test null pointer handling

**Priority:** HIGH - Foundation for robustness

#### G2. Go: Unicode Edge Cases
**File:** `_src_c/go/main_test.go`

- Test ASCII-only input: `"hello world"`
- Test emoji sequences: `"👍🎉"`
- Test mixed scripts: `"Hello こんにちは 你好"`
- Test surrogate pairs and combining characters
- Test RTL text: `"العربية"`, `"עברית"`
- Test zero-width characters

**Priority:** HIGH - Japanese NLP must handle Unicode correctly

#### G3. Go: Memory & Allocation Critical Cases
**File:** `_src_c/go/main_test.go`

- Test large token arrays (1000+ tokens per input)
- Test repeated allocation/deallocation cycles (stress test)
- Test memory cleanup on error paths
- Test string field allocation with unicode characters (multi-byte)

**Priority:** CRITICAL - Memory leaks would be fatal

#### G4. Go: Concurrency Edge Cases
**File:** `_src_c/go/main_test.go`

- Test high concurrency (1000+ goroutines)
- Test concurrent initialization/destruction of instances
- Test concurrent access to same instance (thread safety verification)
- Test handles persisting across concurrent operations

**Priority:** MEDIUM - Current test only checks map access

#### G5. Go: Error Recovery & State Integrity
**File:** `_src_c/go/main_test.go`

- Test tokenization after failed operations
- Test instance reuse after errors
- Test handle reuse patterns
- Test cleanup of partially-allocated tokens

**Priority:** MEDIUM - Ensures state consistency

---

### PYTHON TESTS: Edge Cases & Robustness

#### P1. Python: Edge Cases - Empty & Boundary Inputs
**File:** `tests/libkagome_test.py` (migrate to pytest)

- Test empty string: `kagome.tokenize("")`
- Test single character: `kagome.tokenize("a")`
- Test very long text (100KB+)
- Test whitespace-only: `kagome.tokenize("   \n\t  ")`
- Test single ASCII character

**Priority:** HIGH - Basic input validation

#### P2. Python: Unicode Edge Cases
**File:** `tests/libkagome_test.py`

- Test ASCII input: `kagome.tokenize("hello world")`
- Test emoji: `kagome.tokenize("👍 Great! 🎉")`
- Test mixed scripts: `kagome.tokenize("English 日本語 中文 العربية")`
- Test combining diacritics: `kagome.tokenize("café naïve")`
- Test surrogate pairs
- Test RTL text

**Priority:** HIGH - International text handling

#### P3. Python: wakati() Method Coverage
**File:** `tests/libkagome_test.py`

- Test empty string: `kagome.wakati("")`
- Test single word: `kagome.wakati("テスト")`
- Test sentence breakdown: `kagome.wakati("今日は天気です")`
- Test ASCII: `kagome.wakati("hello world")`
- Test mixed content
- Verify wakati vs tokenize consistency

**Priority:** HIGH - wakati() currently untested

#### P4. Python: Token Equality & Comparison Edge Cases
**File:** `tests/libkagome_test.py`

- Test Token.__eq__ with None: `token == None`
- Test Token.__eq__ with string: `token == "すもも"`
- Test Token.__eq__ with different Token instances from same text
- Test Token.__eq__ with different texts
- Test Token in list: `token in [token1, token2]`
- Test repr() output parsing (self-consistency)

**Priority:** MEDIUM - New __eq__/__repr__ methods need coverage

#### P5. Python: Multiple Instances & Reuse
**File:** `tests/libkagome_test.py`

- Test creating multiple Kagome instances
- Test sharing instances between threads (current behavior)
- Test instance reuse after multiple tokenizations
- Test memory behavior with many instances

**Priority:** MEDIUM - Real-world usage pattern

#### P6. Python: Version & Metadata
**File:** `tests/libkagome_test.py`

- Test `__version__` is defined and matches pyproject.toml
- Test `__all__` exports Kagome and Token only
- Test module docstring exists
- Test public API accessibility

**Priority:** LOW - Metadata integrity

---

## Deferred Items (v2.1+)

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

---

## Implementation Notes

**Testing Strategy:**
1. Go tests establish correctness and safety at FFI boundary
2. Python tests verify language binding behavior
3. Edge cases discovered in one layer inform the other
4. Priority order: CRITICAL → HIGH → MEDIUM → LOW

**Test Organization:**
- Go: Add to `_src_c/go/main_test.go`
- Python: Create `tests/test_libkagome.py` (pytest format)
- Keep existing `tests/libkagome_test.py` as integration test

**Running Tests:**
```bash
# Go
cd _src_c/go && go test -v ./...

# Python (existing)
PYTHONPATH=./src python3 tests/libkagome_test.py

# Python (new pytest)
pytest tests/test_libkagome.py -v
```
