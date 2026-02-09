# TODO & Issues

Issues and improvements found during code review (2026-02-09).

## Priority Summary for v2.10.3 Initial Release

### MUST FIX BEFORE RELEASE (3 remaining items)

1. ~~Fix `macos-12` → `macos-15-intel` in publish.yml (#1)~~ [DONE - c9ea815]
2. ~~Single-source version (remove duplication #2)~~ [DONE - 1ddddf4]
3. Add `__eq__` / `__repr__` to Token (#3)
4. Add Go tests to publish.yml (#5)
5. Fix Windows Arm64 claim in README (#18)

### SHOULD CONSIDER (5 items)

- Verify macOS deployment target (#7)
- Document thread-safety (#9)
- Add `wakati()` test (#11)
- Add `-pthread` on Linux (#15)
- Python version matrix in CI (#6)

### DEFER TO v2.1+ (9 items)

- Refactor build steps into reusable workflow (#4)
- Context manager support (#8)
- Migrate to pytest (#12)
- Edge-case tests (#13)
- And 5 nice-to-haves (#21-24)

---

## Critical / Bugs

### 1. publish.yml still uses deprecated `macos-12` runner [BLOCKING]

[publish.yml](.github/workflows/publish.yml) line 33 still has `os: macos-12`, which was already fixed in `build-and-test.yml` → `macos-15-intel`. The publish workflow will hang indefinitely on release.

**File:** `.github/workflows/publish.yml`
**Fix:** Change `macos-12` → `macos-15-intel` (same fix applied to `build-and-test.yml`)
**Priority:** CRITICAL - Must fix before first release

### 2. Version is duplicated in two files [IMPORTANT]

`__version__` is defined in both `pyproject.toml` and `src/libkagome/__init__.py`. They can easily drift.

**Files:** `pyproject.toml`, `src/libkagome/__init__.py`
**Fix:** Use `importlib.metadata.version("kagome-py")` in `__init__.py` to read from installed metadata
**Priority:** HIGH - Maintenance issue, should fix before first release

### 3. `Token` class lacks `__eq__` and `__repr__` [SHOULD FIX]

`Token.__str__` exists but `__eq__` and `__repr__` are missing. This makes testing and debugging harder.

**File:** `src/libkagome/_wrapper.py`
**Fix:** Use `@dataclass(frozen=True)` or add `__eq__`, `__hash__`, and `__repr__` methods
**Priority:** MEDIUM - Code quality and testing, should fix for initial release

---

## CI/CD

### 4. Build steps are duplicated between `build-and-test.yml` and `publish.yml` [REFACTOR]

The two workflows have near-identical build steps. Any fix to one must be manually applied to the other.

**Fix:** Extract shared build steps into a reusable workflow (`workflow_call`) or a composite action
**Priority:** MEDIUM - Refactoring, can defer to future (but should do to prevent drift)

### 5. Go tests are skipped in `publish.yml` [SHOULD FIX]

`build-and-test.yml` runs `go test -v ./...` but `publish.yml` skips it. A release could be published from untested Go code.

**File:** `.github/workflows/publish.yml`
**Fix:** Add `go test -v ./...` step, or extract into reusable workflow (see #4)
**Priority:** HIGH - Ensure released code is tested

### 6. No Python version matrix in CI [ENHANCEMENT]

CI only tests against Python 3.12, but `pyproject.toml` declares 3.10–3.13 support.

**Fix:** Add test matrix for Python 3.10/3.11/3.12/3.13
**Priority:** MEDIUM - Testing, important for multi-version support

### 7. `wheel_plat` tag `macosx_10_13_x86_64` may be too low [VERIFY]

Wheel built on macOS 15 but tagged as `macosx_10_13`. Deployment target may be incompatible.

**Fix:** Verify actual minimum macOS version requirement, or bump to `macosx_13_0_x86_64`
**Priority:** MEDIUM - May cause runtime failures on older macOS

---

## Python Code

### 8. `Kagome` does not support context manager protocol [ENHANCEMENT]

Users must rely on `__del__` for cleanup. A context manager would give deterministic resource cleanup.

**File:** `src/libkagome/_wrapper.py`
**Fix:** Add `__enter__` / `__exit__` methods
**Priority:** LOW - Nice-to-have, can defer to v2.1.0

### 9. `Kagome` is not thread-safe on the Python side [DOCUMENT]

Go side is thread-safe but Python calls aren't protected. GIL masks this in CPython, but it's fragile.

**File:** `src/libkagome/_wrapper.py`
**Fix:** Document thread-safety guarantees, or add `threading.Lock` around FFI calls
**Priority:** MEDIUM - Should at least document current behavior

### 10. Module-level docstring placement [SHOULD FIX - EASY]

Module docstring in `_wrapper.py` is **after** imports, won't be recognized as `__doc__`.

**File:** `src/libkagome/_wrapper.py`
**Fix:** Move docstring before `from __future__` import
**Priority:** LOW - Code quality, easy fix

### 11. `wakati()` test coverage missing [SHOULD ADD]

Test script only covers `tokenize()`, not `wakati()`.

**File:** `tests/libkagome_test.py`
**Fix:** Add `wakati()` test cases
**Priority:** MEDIUM - Test coverage

---

## Testing

### 12. Test script is not using a proper test framework [REFACTOR]

Hand-rolled test script comparing string representations. Fragile and no `pytest` integration.

**Fix:** Migrate to `pytest` with proper assertions
**Priority:** MEDIUM - Better testing infrastructure for future maintenance
**Benefit:** Per-test granularity, coverage reporting, CI integration

### 13. No edge-case tests [ENHANCEMENT]

No tests for empty strings, ASCII, very long input, Unicode edge cases, multiple instances, etc.

**Fix:** Add edge-case test suite
**Priority:** LOW - Can defer to future maintenance

### 14. Go concurrent test doesn't test C interop [KNOWN LIMITATION]

`TestKagomeTokenizeConcurrent` only checks map access, not actual tokenization.

**File:** `_src_c/go/main_test.go`
**Fix:** Document or defer (requires integration test infrastructure)
**Priority:** LOW - Already documented in code

---

## Build / Packaging

### 15. `setup.py` sdist build doesn't pass `-pthread` on Linux [SHOULD FIX]

Missing `-pthread` for Linux. Go runtime requires pthreads. May work by accident but not guaranteed.

**Files:** `setup.py`, `.github/workflows/build-and-test.yml`
**Fix:** Add `-lpthread` (or `-pthread`) to Linux `cc` command
**Priority:** HIGH - Potential runtime failure on some Linux systems

### 16. No `.gitkeep` verification [MINOR]

`.gitignore` references `.gitkeep` files but no CI verification they exist.

**Priority:** LOW - Cosmetic

### 17. MANIFEST.in comment about build artifacts [DOCUMENTATION]

No comment explaining why `_src_c/build/` is excluded yet `libkagome.h` is generated during install.

**File:** `MANIFEST.in`
**Fix:** Add clarifying comment
**Priority:** LOW - Nice to document for future maintainers

---

## Documentation

### 18. README claims Windows Arm64 support [SHOULD FIX - EASY]

README says "Windows (x86_64, Arm64)" but currently only x86_64 is supported.

**Files:** `README.md`, `DEVELOPMENT.md`
**Fix:** Remove Arm64 from README Windows support claim
**Priority:** HIGH - Documentation accuracy for initial release

### 19. Stale documentation files [ALREADY DONE]

Removed `IMPLEMENTATION_SUMMARY.md`, `CI_CD_FIXES.md`, `todo.md` in cleanup pass.
Renamed `QUICKSTART_PIP.md` → `DEVELOPMENT.md`

**Status:** COMPLETE

### 20. `CONTRIBUTING.md` references non-existent `docker/` directory [ALREADY DONE]

Removed reference to non-existent `docker/` directory.

**Status:** COMPLETE

---

## Nice-to-Have (v2.1+)

### 21. Add `py.typed` marker

For downstream type-checking support (`mypy`, `pyright`), add `src/libkagome/py.typed`.

**Priority:** LOW - Future improvement for type hint support

### 22. Add `__all__` to `_wrapper.py`

Make public API explicit. Currently everything is importable.

**Priority:** LOW - Code organization, can defer

### 23. Refactor `Token` to use `@dataclass`

`Token` would get `__eq__`, `__repr__`, `__hash__` automatically (related to #3).

**Priority:** LOW - Would fix #3 simultaneously

### 24. Add CLI entry point

`python -m libkagome "すもも..."` for quick testing.

**Priority:** LOW - Nice-to-have for testing
